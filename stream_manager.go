package main

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/pion/webrtc/v3"
    "github.com/pion/rtp"
)

type StreamManager struct {
    config         StreamConfig
    audioProcessor *AudioProcessor
    videoProcessor *VideoProcessor
    rtmpClient     *RTMPClient
    stats          *StreamStats
    mu             sync.RWMutex

    // Stream state
    isLive         bool
    startTime      time.Time
    viewers        int
    frameBuffer    *FrameBuffer
}

type StreamConfig struct {
    RTMPEndpoint    string
    StreamKey       string
    VideoCodec      string
    AudioCodec      string
    VideoBitrate    int
    AudioBitrate    int
    FrameRate       int
    Resolution      Resolution
    KeyframeInterval int
}

type StreamStats struct {
    StartTime       time.Time
    Duration        time.Duration
    BytesSent      uint64
    FramesSent     uint64
    PacketsLost    uint32
    Bitrate        float64
    ViewerCount    int
    Health         float64
}

type Resolution struct {
    Width  int
    Height int
}

func NewStreamManager(config StreamConfig) (*StreamManager, error) {
    sm := &StreamManager{
        config:      config,
        frameBuffer: NewFrameBuffer(config.FrameRate),
        stats:       &StreamStats{},
    }

    // Initialize processors
    var err error
    sm.audioProcessor, err = NewAudioProcessor(config.AudioCodec, config.AudioBitrate)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize audio processor: %w", err)
    }

    sm.videoProcessor, err = NewVideoProcessor(config.VideoCodec, config.VideoBitrate, config.Resolution)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize video processor: %w", err)
    }

    // Initialize RTMP client
    sm.rtmpClient, err = NewRTMPClient(config.RTMPEndpoint, config.StreamKey)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize RTMP client: %w", err)
    }

    return sm, nil
}

func (sm *StreamManager) StartStream(ctx context.Context) error {
    sm.mu.Lock()
    if sm.isLive {
        sm.mu.Unlock()
        return fmt.Errorf("stream already running")
    }

    sm.isLive = true
    sm.startTime = time.Now()
    sm.mu.Unlock()

    // Start stream components
    errCh := make(chan error, 3)
    
    go func() {
        errCh <- sm.audioProcessor.Start(ctx)
    }()
    
    go func() {
        errCh <- sm.videoProcessor.Start(ctx)
    }()
    
    go func() {
        errCh <- sm.rtmpClient.Connect(ctx)
    }()

    // Monitor stream health
    go sm.monitorStream(ctx)

    // Start main stream loop
    go sm.streamLoop(ctx)

    // Handle errors
    select {
    case err := <-errCh:
        sm.StopStream()
        return fmt.Errorf("stream component failed: %w", err)
    case <-ctx.Done():
        sm.StopStream()
        return ctx.Err()
    }
}

func (sm *StreamManager) StopStream() error {
    sm.mu.Lock()
    defer sm.mu.Unlock()

    if !sm.isLive {
        return nil
    }

    sm.isLive = false
    
    // Stop all components
    sm.audioProcessor.Stop()
    sm.videoProcessor.Stop()
    sm.rtmpClient.Disconnect()

    return nil
}

func (sm *StreamManager) streamLoop(ctx context.Context) {
    ticker := time.NewTicker(time.Second / time.Duration(sm.config.FrameRate))
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := sm.processFrame(); err != nil {
                log.Printf("Frame processing error: %v", err)
            }
        }
    }
}

func (sm *StreamManager) processFrame() error {
    // Get next frame from buffer
    frame := sm.frameBuffer.NextFrame()
    
    // Process video frame
    videoPacket, err := sm.videoProcessor.ProcessFrame(frame)
    if err != nil {
        return fmt.Errorf("video processing failed: %w", err)
    }

    // Process audio
    audioPacket, err := sm.audioProcessor.ProcessFrame()
    if err != nil {
        return fmt.Errorf("audio processing failed: %w", err)
    }

    // Send to RTMP
    if err := sm.rtmpClient.SendPackets(videoPacket, audioPacket); err != nil {
        return fmt.Errorf("failed to send packets: %w", err)
    }

    // Update stats
    sm.updateStats(videoPacket, audioPacket)

    return nil
}

func (sm *StreamManager) updateStats(video, audio []byte) {
    sm.mu.Lock()
    defer sm.mu.Unlock()

    sm.stats.BytesSent += uint64(len(video) + len(audio))
    sm.stats.FramesSent++
    sm.stats.Duration = time.Since(sm.stats.StartTime)
    sm.stats.Bitrate = float64(sm.stats.BytesSent*8) / sm.stats.Duration.Seconds()
}

func (sm *StreamManager) monitorStream(ctx context.Context) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            sm.checkStreamHealth()
        }
    }
}

func (sm *StreamManager) checkStreamHealth() {
    sm.mu.Lock()
    defer sm.mu.Unlock()

    //
} 