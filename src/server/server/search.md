# Search Query Syntax Guide

## Overview

This document describes the search query syntax used in our search system. The search engine supports various logical operators, wildcards, and grouping to help you find exactly what you're looking for.

## Logical Operators

| Operator | Symbol | Description | Usage Example |
|----------|--------|-------------|---------------|
| AND      | `AND` or space | Matches documents containing all specified terms | `cat AND dog` or `cat dog` |
| OR       | `OR` or `\|` | Matches documents containing any of the specified terms | `cat OR dog` or `cat \| dog` |
| NOT      | `NOT` or `!` | Excludes documents containing the specified term | `cat NOT dog` or `cat AND !dog` |
| Grouping | `()` | Groups terms for complex logical expressions | `(cat OR dog) AND mouse` |

## Search Term Formats

### 1. Basic Terms
- **Single Word**: `hello`
- [TBD] **Phrase Search**: `"hello world"` (use double quotes for exact phrase matching)
- **Case Sensitivity**: Search is case-insensitive by default

### 2. Wildcard Matching
- **Prefix Matching**: `hello*` (matches "hello", "hello2", "helloworld")
- **Multiple Wildcards**: `hel*o` (matches "hello", "helio", "hell ok")
  - **Note**: Wildcards can only be used at the end of a word or between characters, and at least 2 characters in prefix
    - `*ello` - invalid!
    - `h*ll` - invalid!
    - `he*l` - valid
    - `he*` - valid

### 3. Special Characters
- **Quotes**: `"exact phrase"` for exact matching
- **Parentheses**: `()` for grouping
- **Escape Character**: Use `\` to escape special characters

## Examples

### 1. Single Word Search
- `hello` → matches: "hello", "Hello", "HELLO", "hello world"
- `"hello world"` → matches: "hello world" (exact phrase only)

### 2. AND Search
- `cat dog` → matches documents containing both "cat" and "dog"
- `cat* AND dog*` → matches: "cat dog", "cats dogs", "catalog doggy"
- `"cat food" AND "dog food"` → matches documents containing both exact phrases

### 3. OR Search
- `cat OR dog` → matches documents containing either "cat" or "dog"
- `cat* OR dog*` → matches: "cat", "cats", "dog", "dogs"
- `(cat OR dog) AND food` → matches: "cat food", "dog food"

### 4. NOT Search
- `cat NOT dog` → matches documents with "cat" but not "dog"
- `cat AND NOT dog` → same as above
- `(cat OR dog) AND NOT mouse` → matches documents with "cat" or "dog" but not "mouse"

### 5. Complex Expressions
- `(cat OR dog) AND (food OR treat)` → matches: "cat food", "dog treat", "cat treat"
- `(cat* OR dog*) AND NOT (mouse OR rat)` → matches documents with words starting with "cat" or "dog" but not containing "mouse" or "rat"

## Performance Tips

1. **Use Specific Terms**
   - Prefer specific terms over wildcards when possible
   - Example: Use `cat` instead of `ca*` when you know the exact term

2. **Optimize NOT Queries**
   - Always combine NOT with AND to avoid performance issues
   - Bad: `NOT cat`
   - Good: `dog AND NOT cat`

3. **Phrase Search**
   - Use phrase search for exact matches
   - Example: `"cat food"` instead of `cat AND food`

4. **Grouping**
   - Use parentheses to make complex queries more efficient
   - Example: `(cat OR dog) AND (food OR treat)`

## Common Use Cases

### 1. Finding Related Documents
```
(cat OR dog) AND (food OR treat) AND NOT (mouse OR rat)
```

### 2. Exact Phrase Matching
```
"cat food" AND "dog food"
```

### 3. Wildcard Search
```
cat* AND food*
```

### 4. Excluding Specific Terms
```
animal AND NOT (mouse OR rat)
```

## Troubleshooting

1. **No Results Found**
   - Check for typos
   - Try removing wildcards
   - Use simpler terms

2. **Too Many Results**
   - Add more specific terms
   - Use phrase search
   - Add NOT conditions

3. **Performance Issues**
   - Avoid single NOT queries
   - Limit wildcard usage
   - Use more specific terms
