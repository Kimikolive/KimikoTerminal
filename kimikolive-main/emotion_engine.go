package main

import (
    "context"
    "math"
    "sync"
    "time"

    "github.com/sashabaranov/go-openai"
)

type EmotionEngine struct {
    model           string
    client          *openai.Client
    currentEmotion  string
    LastConfidence  float64
    emotionHistory  []EmotionRecord
    mu             sync.RWMutex
    
    // Emotional state parameters
    arousal         float64
    valence         float64
    dominance       float64
}

type EmotionRecord struct {
    Emotion    string
    Timestamp  time.Time
    Confidence float64
    Source     string
}

type EmotionState struct {
    Primary    string
    Secondary  string
    Intensity  float64
    Valence    float64
    Arousal    float64
    Dominance  float64
}

func NewEmotionEngine(model string) *EmotionEngine {
    return &EmotionEngine{
        model:          model,
        emotionHistory: make([]EmotionRecord, 0, 100),
        arousal:        0.5,
        valence:        0.5,
        dominance:      0.5,
    }
}

func (e *EmotionEngine) AnalyzeEmotion(text string) (string, float64) {
    e.mu.Lock()
    defer e.mu.Unlock()

    // Use OpenAI to analyze emotion
    completion, err := e.client.CreateCompletion(context.Background(), openai.CompletionRequest{
        Model:       e.model,
        Prompt:      generateEmotionPrompt(text),
        MaxTokens:   50,
        Temperature: 0.3,
    })
    if err != nil {
        return "neutral", 0.5
    }

    // Parse emotion and confidence
    emotion, confidence := parseEmotionResponse(completion.Choices[0].Text)
    
    // Update emotional state
    e.updateEmotionalState(emotion, confidence)
    
    return emotion, confidence
}

func (e *EmotionEngine) updateEmotionalState(emotion string, confidence float64) {
    // Update emotion history
    e.emotionHistory = append(e.emotionHistory, EmotionRecord{
        Emotion:    emotion,
        Timestamp:  time.Now(),
        Confidence: confidence,
        Source:     "analysis",
    })

    // Trim history if needed
    if len(e.emotionHistory) > 100 {
        e.emotionHistory = e.emotionHistory[1:]
    }

    // Update VAD (Valence-Arousal-Dominance) values
    vadValues := getVADValues(emotion)
    e.valence = e.valence*0.7 + vadValues.Valence*0.3
    e.arousal = e.arousal*0.7 + vadValues.Arousal*0.3
    e.dominance = e.dominance*0.7 + vadValues.Dominance*0.3
}

func (e *EmotionEngine) GetTemperatureModifier(emotion string) float64 {
    // Calculate temperature modifier based on emotional state
    arousalMod := (e.arousal - 0.5) * 0.2
    valenceMod := (e.valence - 0.5) * 0.1
    
    return arousalMod + valenceMod
}

func (e *EmotionEngine) GetCurrentEmotionalState() EmotionState {
    e.mu.RLock()
    defer e.mu.RUnlock()

    return EmotionState{
        Primary:   e.currentEmotion,
        Secondary: e.calculateSecondaryEmotion(),
        Intensity: e.calculateEmotionalIntensity(),
        Valence:   e.valence,
        Arousal:   e.arousal,
        Dominance: e.dominance,
    }
}

func (e *EmotionEngine) calculateEmotionalIntensity() float64 {
    // Calculate intensity based on VAD values
    return math.Sqrt(
        math.Pow(e.valence-0.5, 2) +
        math.Pow(e.arousal-0.5, 2) +
        math.Pow(e.dominance-0.5, 2),
    ) / math.Sqrt(0.75)
}

func (e *EmotionEngine) calculateSecondaryEmotion() string {
    // Implement complex emotion blending logic
    // This is a simplified version
    if len(e.emotionHistory) < 2 {
        return "neutral"
    }
    
    return e.emotionHistory[len(e.emotionHistory)-2].Emotion
} 