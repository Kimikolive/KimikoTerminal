package main

import (
    "context"
    "encoding/json"
    "math"
    "sync"
    "time"
)

type PersonalitySystem struct {
    baseTraits      PersonalityTraits
    currentState    PersonalityState
    memoryBuffer    *MemoryBuffer
    emotionEngine   *EmotionEngine
    learningRate    float64
    mu              sync.RWMutex

    // Personality adaptation
    traitHistory    []TraitSnapshot
    interactions    []Interaction
    adaptiveRules   map[string]AdaptiveRule
}

type PersonalityTraits struct {
    Openness        float64 `json:"openness"`
    Conscientiousness float64 `json:"conscientiousness"`
    Extraversion    float64 `json:"extraversion"`
    Agreeableness   float64 `json:"agreeableness"`
    Neuroticism     float64 `json:"neuroticism"`
    
    // Additional custom traits
    Playfulness     float64 `json:"playfulness"`
    Creativity      float64 `json:"creativity"`
    Empathy         float64 `json:"empathy"`
    Curiosity       float64 `json:"curiosity"`
    Assertiveness   float64 `json:"assertiveness"`
}

type PersonalityState struct {
    CurrentTraits   PersonalityTraits
    Mood           EmotionState
    Energy         float64
    Engagement     float64
    LastUpdate     time.Time
}

type TraitSnapshot struct {
    Traits    PersonalityTraits
    Timestamp time.Time
    Context   string
}

type Interaction struct {
    UserInput      string
    Response       string
    EmotionalImpact float64
    TraitInfluence  map[string]float64
    Timestamp      time.Time
}

type AdaptiveRule struct {
    Condition      func(PersonalityState, Interaction) bool
    TraitModifiers map[string]float64
    Priority       int
    Cooldown       time.Duration
    LastTriggered  time.Time
}

func NewPersonalitySystem(baseTraits PersonalityTraits) *PersonalitySystem {
    ps := &PersonalitySystem{
        baseTraits:    baseTraits,
        learningRate:  0.01,
        adaptiveRules: initializeAdaptiveRules(),
        traitHistory:  make([]TraitSnapshot, 0),
        interactions:  make([]Interaction, 0),
    }

    ps.currentState = PersonalityState{
        CurrentTraits: baseTraits,
        Energy:       1.0,
        Engagement:   1.0,
        LastUpdate:   time.Now(),
    }

    return ps
}

func (ps *PersonalitySystem) ProcessInteraction(ctx context.Context, input string, response string, emotionalImpact float64) {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    // Record interaction
    interaction := Interaction{
        UserInput:      input,
        Response:       response,
        EmotionalImpact: emotionalImpact,
        TraitInfluence:  ps.calculateTraitInfluence(input, response),
        Timestamp:      time.Now(),
    }
    ps.interactions = append(ps.interactions, interaction)

    // Update personality state
    ps.updatePersonalityState(interaction)

    // Apply adaptive rules
    ps.applyAdaptiveRules(interaction)

    // Take trait snapshot
    ps.takeTraitSnapshot("interaction")
}

func (ps *PersonalitySystem) calculateTraitInfluence(input, response string) map[string]float64 {
    influences := make(map[string]float64)
    
    // Complex trait influence calculation based on interaction content
    // Natural language processing could be used here
    
    return influences
}

func (ps *PersonalitySystem) updatePersonalityState(interaction Interaction) {
    // Update energy and engagement
    timeSinceLastUpdate := time.Since(ps.currentState.LastUpdate)
    energyDecay := math.Exp(-float64(timeSinceLastUpdate) / float64(time.Hour))
    
    ps.currentState.Energy = ps.currentState.Energy*energyDecay + interaction.EmotionalImpact*0.2
    ps.currentState.Engagement = calculateEngagement(ps.interactions)
    
    // Apply trait influences
    for trait, influence := range interaction.TraitInfluence {
        currentValue := ps.getTraitValue(trait)
        newValue := currentValue + influence*ps.learningRate
        ps.setTraitValue(trait, clampTrait(newValue))
    }
    
    ps.currentState.LastUpdate = time.Now()
}

func (ps *PersonalitySystem) applyAdaptiveRules(interaction Interaction) {
    for _, rule := range ps.adaptiveRules {
        if rule.Condition(ps.currentState, interaction) &&
           time.Since(rule.LastTriggered) > rule.Cooldown {
            
            for trait, modifier := range rule.TraitModifiers {
                currentValue := ps.getTraitValue(trait)
                newValue := currentValue + modifier
                ps.setTraitValue(trait, clampTrait(newValue))
            }
            
            rule.LastTriggered = time.Now()
        }
    }
}

func (ps *PersonalitySystem) GeneratePrompt() string {
    ps.mu.RLock()
    defer ps.mu.RUnlock()

    // Generate a detailed personality prompt based on current traits
    traits := ps.currentState.CurrentTraits
    
    return fmt.Sprintf(`You are an AI VTuber with the following personality traits:
- Openness: %.2f (You %s)
- Conscientiousness: %.2f (You %s)
- Extraversion: %.2f (You %s)
- Agreeableness: %.2f (You %s)
- Neuroticism: %.2f (You %s)
- Playfulness: %.2f (You %s)
- Creativity: %.2f (You %s)
- Empathy: %.2f (You %s)
- Curiosity: %.2f (You %s)
- Assertiveness: %.2f (You %s)

Current Energy Level: %.2f
Current Engagement Level: %.2f

Respond in a way that naturally reflects these personality traits.`,
        traits.Openness, describeTraitLevel(traits.Openness),
        traits.Conscientiousness, describeTraitLevel(traits.Conscientiousness),
        traits.Extraversion, describeTraitLevel(traits.Extraversion),
        traits.Agreeableness, describeTraitLevel(traits.Agreeableness),
        traits.Neuroticism, describeTraitLevel(traits.Neuroticism),
        traits.Playfulness, describeTraitLevel(traits.Playfulness),
        traits.Creativity, describeTraitLevel(traits.Creativity),
        traits.Empathy, describeTraitLevel(traits.Empathy),
        traits.Curiosity, describeTraitLevel(traits.Curiosity),
        traits.Assertiveness, describeTraitLevel(traits.Assertiveness),
        ps.currentState.Energy,
        ps.currentState.Engagement,
    )
}

func (ps *PersonalitySystem) takeTraitSnapshot(context string) {
    snapshot := TraitSnapshot{
        Traits:    ps.currentState.CurrentTraits,
        Timestamp: time.Now(),
        Context:   context,
    }
    ps.traitHistory = append(ps.traitHistory, snapshot)
}

func calculateEngagement(interactions []Interaction) float64 {
    if len(interactions) < 2 {
        return 1.0
    }
    
    // Calculate engagement based on interaction frequency and emotional impact
    // Complex engagement calculation logic here
    return 0.0
}

func clampTrait(value float64) float64 {
    return math.Max(0.0, math.Min(1.0, value))
}

func describeTraitLevel(value float64) string {
    switch {
    case value >= 0.8:
        return "strongly exhibit this trait"
    case value >= 0.6:
        return "tend to show this trait"
    case value >= 0.4:
        return "are moderate in this trait"
    case value >= 0.2:
        return "occasionally show this trait"
    default:
        return "rarely display this trait"
    }
} 