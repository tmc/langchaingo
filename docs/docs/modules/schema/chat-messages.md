---
hide_table_of_contents: true
sidebar_position: 1
---

# Chat Messages

The primary interface through which end users interact with LLMs is a chat interface. For this reason, some model providers have started providing access to the underlying API in a way that expects chat messages. These messages have a content field (which is usually text) and are associated with a user (or role). Right now the supported users are System, Human, and AI.

## SystemChatMessage

A chat message representing information that should be instructions to the AI system.

## HumanChatMessage

A chat message representing information coming from a human interacting with the AI system.

## AIChatMessage

A chat message representing information coming from the AI system.

