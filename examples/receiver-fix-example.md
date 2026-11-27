# Example: Fixing Mixed Receiver Types

This document demonstrates how to fix the mixed receiver issue in the `PDF` type from `documentloaders/pdf.go`.

## Problem

The `PDF` type has mixed receivers:

```go
// Pointer receiver (modifies the struct)
func (p *PDF) getPassword() string {
    pass := p.password
    p.password = ""  // Clears password after use
    return pass
}

// Value receiver (doesn't modify struct)
func (p PDF) Load(_ context.Context) ([]schema.Document, error) {
    // ...
    if p.password != "" {
        reader, err = pdf.NewReaderEncrypted(p.r, p.s, p.getPassword)
        //                                              ^^^^^^^^^^^^
        // BUG: This calls getPassword on a copy, so original password is never cleared!
    }
    // ...
}
```

## Issue Analysis

1. `getPassword()` uses pointer receiver to modify the struct (clear password after use)
2. `Load()` uses value receiver, so `p` is a copy of the original struct
3. When `Load()` calls `p.getPassword()`, it passes the address of the copy
4. The password is cleared on the copy, not the original struct
5. This breaks the intended "use password once" security feature

## Solution

Convert all methods to use pointer receivers for consistency:

```go
// Keep pointer receiver (correct)
func (p *PDF) getPassword() string {
    pass := p.password
    p.password = ""
    return pass
}

// Change to pointer receiver
func (p *PDF) Load(_ context.Context) ([]schema.Document, error) {
    // Now p.getPassword() works correctly
    if p.password != "" {
        reader, err = pdf.NewReaderEncrypted(p.r, p.s, p.getPassword)
    }
    // ...
}

// Change to pointer receiver  
func (p *PDF) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
    docs, err := p.Load(ctx)
    // ...
}
```

## Breaking Changes

This change affects:
1. **Interface compatibility**: The `Loader` interface expects value receivers
2. **User code**: Any code that calls these methods directly

## Migration Strategy

### Option 1: Update Interface (Preferred)
```go
// Update the Loader interface to use pointer receivers
type Loader interface {
    Load(ctx context.Context) ([]schema.Document, error)
    LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error)
}

// Update all implementations consistently
```

### Option 2: Fix Without Interface Changes
```go
// Alternative: Modify getPassword to work with value receivers
func (p PDF) getPasswordOnce() (string, PDF) {
    newP := p
    newP.password = ""
    return p.password, newP
}

func (p PDF) Load(_ context.Context) ([]schema.Document, error) {
    if p.password != "" {
        password, _ := p.getPasswordOnce()
        reader, err = pdf.NewReaderEncrypted(p.r, p.s, func() string { return password })
    }
    // ...
}
```

### Option 3: Stateless Design (Best)
```go
// Even better: Make the password an explicit parameter
func (p PDF) LoadWithPassword(_ context.Context, password string) ([]schema.Document, error) {
    // Use password directly without storing it
}

// Keep Load() for backward compatibility
func (p PDF) Load(ctx context.Context) ([]schema.Document, error) {
    return p.LoadWithPassword(ctx, p.password)
}
```

## Recommendation

For this specific case, **Option 3** is best because:
1. More secure (password isn't stored in struct)
2. Thread-safe (no state mutation)
3. Clearer API (explicit about password usage)
4. Maintains backward compatibility

This demonstrates why receiver analysis is important - it can reveal actual bugs, not just style issues!