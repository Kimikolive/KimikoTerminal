package main

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "sync"

    "github.com/sashabaranov/go-openai"
)

type LLMProcessor struct {
    client         *openai.Client
    config         AIConfig
    memoryBuffer   *MemoryBuffer
    emotionEngine  *EmotionEngine
    personality    *PersonalityVector
    mu            sync.Mutex
    
    // Conversation state
    contextWindow  []Message
    lastResponse   time.Time
    emotionState   EmotionState
    interactionCount int
}

type Message struct {
    Role      string    `json:"role"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
    Emotion   string    `json:"emotion"`
    Confidence float64  `json:"confidence"`
}

func NewLLMProcessor(config AIConfig, openAIKey string) *LLMProcessor {
    return &LLMProcessor{
        client: openai.NewClient(openAIKey),
        config: config,
        memoryBuffer: NewMemoryBuffer(config.MemoryBufferSize),
        emotionEngine: NewEmotionEngine(config.EmotionModel),
        personality: NewPersonalityVector(config.PersonalityVector),
        contextWindow: make([]Message, 0, config.ContextWindowSize),
    }
}

func (l *LLMProcessor) ProcessInput(ctx context.Context, input string) (*Response, error) {
    l.mu.Lock()
    defer l.mu.Unlock()

    // Analyze input emotion
    emotion, confidence := l.emotionEngine.AnalyzeEmotion(input)
    
    // Build context with personality injection
    messages := l.buildContextMessages()
    messages = append(messages, Message{
        Role:      "user",
        Content:   input,
        Timestamp: time.Now(),
        Emotion:   emotion,
        Confidence: confidence,
    })

    // Generate response with dynamic temperature
    temp := l.calculateDynamicTemperature(emotion, confidence)
    
    resp, err := l.client.CreateChatCompletion(
        ctx,
        openai.ChatCompletionRequest{
            Model:       l.config.Model,
            Messages:    l.convertToOpenAIMessages(messages),
            Temperature: temp,
            MaxTokens:   1000,
            TopP:        0.9,
            PresencePenalty: 0.6,
            FrequencyPenalty: 0.3,
        },
    )
    if err != nil {
        return nil, fmt.Errorf("LLM processing error: %w", err)
    }

    // Process response
    response := &Response{
        Text:     resp.Choices[0].Message.Content,
        Emotion:  l.emotionEngine.AnalyzeResponse(resp.Choices[0].Message.Content),
        Metadata: l.generateResponseMetadata(),
    }

    // Update memory and context
    l.updateMemoryAndContext(response)
    
    return response, nil
}

func (l *LLMProcessor) buildContextMessages() []Message {
    var messages []Message
    
    // Add personality base prompt
    messages = append(messages, l.personality.GenerateBasePrompt())
    
    // Add relevant memories
    memories := l.memoryBuffer.GetRelevantMemories(l.contextWindow)
    messages = append(messages, memories...)
    
    // Add recent context
    messages = append(messages, l.contextWindow...)
    
    return messages
}

func (l *LLMProcessor) calculateDynamicTemperature(emotion string, confidence float64) float64 {
    baseTemp := l.config.TemperatureBase
    
    // Adjust temperature based on emotion and confidence
    emotionMod := l.emotionEngine.GetTemperatureModifier(emotion)
    confidenceMod := 0.2 * (1 - confidence)
    
    // Add some randomness for variety
    randomMod := (time.Now().UnixNano() % 100) / 1000.0
    
    return baseTemp + emotionMod + confidenceMod + randomMod
}

func (l *LLMProcessor) updateMemoryAndContext(response *Response) {
    // Update context window
    l.contextWindow = append(l.contextWindow, Message{
        Role:      "assistant",
        Content:   response.Text,
        Timestamp: time.Now(),
        Emotion:   response.Emotion,
        Confidence: 1.0,
    })
    
    // Trim context if needed
    if len(l.contextWindow) > l.config.ContextWindowSize {
        l.contextWindow = l.contextWindow[1:]
    }
    
    // Update memory buffer
    l.memoryBuffer.AddMemory(response)
    
    // Update interaction count
    l.interactionCount++
}

type Response struct {
    Text     string
    Emotion  string
    Metadata ResponseMetadata
}

type ResponseMetadata struct {
    Timestamp       time.Time
    InteractionNum  int
    ContextSize     int
    Temperature     float64
    EmotionConfidence float64
}

func (l *LLMProcessor) generateResponseMetadata() ResponseMetadata {
    return ResponseMetadata{
        Timestamp:       time.Now(),
        InteractionNum:  l.interactionCount,
        ContextSize:     len(l.contextWindow),
        Temperature:     l.config.TemperatureBase,
        EmotionConfidence: l.emotionEngine.LastConfidence,
    }
} 