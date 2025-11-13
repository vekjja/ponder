# Ponder

**OpenAI-Powered Multi-Modal Chat Tool**

[![main](https://github.com/seemywingz/ponder/actions/workflows/dockerBuildX.yml/badge.svg?branch=v0.4.2)](https://github.com/seemywingz/ponder/actions/workflows/dockerBuildX.yml)

Ponder is a versatile command-line application and Discord bot that leverages OpenAI's API to provide interactive chat, image generation, text-to-speech, and immersive text adventures. Built with Go, it offers a beautiful terminal UI powered by Bubble Tea and supports both local CLI usage and containerized deployments.

---

## 🚀 Features

- **Interactive Chat** - Beautiful TUI chat interface with conversation history
- **Discord Bot Integration** - Deploy as a Discord bot with slash commands and message responses
- **Image Generation** - Create images using DALL-E 3 with download and auto-open options
- **Text-to-Speech** - Convert text to speech with multiple voice options
- **Text Adventure Mode** - Immersive, AI-driven text adventure game with character customization
- **Voice Narration** - Optional audio narration for chat responses
- **Configurable** - Extensive YAML configuration with sensible defaults
- **Kubernetes Ready** - Includes Helm charts for production deployments

---

## 📦 Installation

### Via Go Install
```bash
go install github.com/vekjja/ponder@latest
```

### Via Docker
```bash
docker pull ghcr.io/vekjja/ponder:latest
```

### Via Helm (Kubernetes)
```bash
helm upgrade --install ponder ./helm --values ./helm/values.production.yaml
```

---

## ⚙️ Configuration

### Required Environment Variables

```bash
# Required for OpenAI functionality
OPENAI_API_KEY={YOUR_OPENAI_API_KEY}

# Required only for Discord bot
DISCORD_API_KEY={YOUR_DISCORD_BOT_API_KEY}
```

**Getting API Keys:**
- **OpenAI**: Visit [OpenAI API Keys](https://platform.openai.com/account/api-keys)
- **Discord**: Visit [Discord Developer Portal](https://discord.com/developers/applications)

### Optional Configuration File

Ponder supports YAML configuration files for advanced settings. Place a `config` file in one of these locations:
- Current directory (`./config`)
- `./files/config`
- `$HOME/.ponder/config`

Or specify a custom location with `--config` flag.

**Example Configuration:**
```yaml
openAI_endpoint: "https://api.openai.com/v1/"
openAI_chat_model: "gpt-4"
openAI_image_model: "dall-e-3"
openAI_image_size: "1024x1024"
openAI_image_downloadPath: "~/Ponder/Images/"
openAI_tts_model: "tts-1"
openAI_tts_voice: "onyx"
openAI_chat_systemMessage: "You are a helpful assistant."
```

---

## 🎯 Usage

### Quick Chat
Ask a single question:
```bash
ponder "What is artificial intelligence?"
```

### Interactive Chat Mode
Start a conversational session:
```bash
ponder chat
```

Or use the root command:
```bash
ponder
```

### Image Generation
Generate images with DALL-E 3:
```bash
# Basic usage
ponder image "a majestic dragon flying over mountains at sunset"

# Download image to local directory
ponder image "futuristic cityscape" --download

# Open image in system viewer
ponder image "abstract art" --open

# Generate multiple images
ponder image "cute robot" --count 3
```

**Image Options:**
- `-d, --download` - Download image(s) to configured directory (default: `~/Ponder/Images/`)
- `-o, --open` - Automatically open image in system default viewer
- `-c, --count N` - Generate N images (default: 1)

### Text-to-Speech
Convert text to speech:
```bash
# Interactive TTS mode
ponder tts

# Direct conversion with narration
ponder "Tell me a story" --narrate

# Save audio to file
ponder tts --file output.mp3

# Use different voice
ponder tts --voice nova
```

**Available Voices:**
`alloy`, `ash`, `coral`, `echo`, `fable`, `onyx`, `nova`, `sage`, `shimmer`

### Text Adventure
Dive into an AI-powered text adventure:
```bash
# Start adventure
ponder adventure

# Adventure with image generation
ponder adventure --images

# Adventure with narration
ponder adventure --narrate
```

The adventure mode will:
1. Prompt you for your character name
2. Ask for a character description
3. Generate a dynamic story based on your choices
4. Track character stats (HP, MP, Level, Strength, Defense, Dexterity, Intellect, Hunger)

### Discord Bot
Run Ponder as a Discord bot:
```bash
ponder discord-bot
```

**Discord Features:**
- Responds to direct messages
- Responds to @mentions in channels
- `/ponder-image` slash command for image generation
- Context-aware conversations (remembers recent messages)

**Deregister Discord Commands:**
```bash
ponder discord-bot --deregister-commands "command_id_1,command_id_2"
```

---

## 🐳 Docker Usage

### Single Query
```bash
docker run -e OPENAI_API_KEY=$OPENAI_API_KEY \
  ghcr.io/vekjja/ponder:latest "What is quantum computing?"
```

### Discord Bot
```bash
docker run -e OPENAI_API_KEY=$OPENAI_API_KEY \
  -e DISCORD_API_KEY=$DISCORD_API_KEY \
  ghcr.io/vekjja/ponder:latest discord-bot
```

### With Custom Config
```bash
docker run -e OPENAI_API_KEY=$OPENAI_API_KEY \
  -v $(pwd)/config:/app/config \
  ghcr.io/vekjja/ponder:latest --config /app/config
```

---

## ☸️ Kubernetes Deployment

Ponder includes production-ready Helm charts for Kubernetes deployment.

### Deploy Discord Bot to Kubernetes
```bash
# Create secrets
kubectl create secret generic openai-api-key --from-literal=api-key=YOUR_OPENAI_KEY
kubectl create secret generic discord-api-key --from-literal=api-key=YOUR_DISCORD_KEY

# Deploy with Helm
helm upgrade --install ponder ./helm --values ./helm/values.production.yaml
```

**Helm Configuration:**
The chart supports:
- Configurable replicas and autoscaling
- Resource limits and requests
- Ingress configuration
- Custom configuration via ConfigMap
- Environment-specific values files (development/production)

---

## 🎨 Global Flags

These flags work with most commands:

```bash
-v, --verbose          Verbose output (use -v, -vv, -vvv for increased verbosity)
-n, --narrate          Narrate responses using TTS
    --voice string     TTS voice (default "onyx")
    --config string    Path to config file
-h, --help             Help for any command
```

---

## 📚 Commands Reference

```
Available Commands:
  adventure   Interactive text adventure game
  chat        Open-ended chat with OpenAI
  completion  Generate shell autocompletion scripts
  discord-bot Run as Discord bot
  help        Help about any command
  image       Generate images from text prompts
  tts         Text-to-Speech conversion
```

Get detailed help for any command:
```bash
ponder [command] --help
```

---

## 🛠️ Development

### Project Structure
```
ponder/
├── cmd/                    # Command implementations
│   ├── adventure.go       # Text adventure logic
│   ├── chat.go            # Chat functionality
│   ├── discord.go         # Discord bot command
│   ├── Discord_api.go     # Discord API handlers
│   ├── image.go           # Image generation
│   ├── tts.go             # Text-to-speech
│   ├── chatHistoryModel.go # Bubble Tea UI models
│   ├── root.go            # Root command and config
│   └── utils.go           # Utility functions
├── helm/                   # Kubernetes Helm charts
├── files/                  # Default config files
├── go.mod                  # Go module definition
└── main.go                 # Application entry point
```

### Building from Source
```bash
# Clone repository
git clone https://github.com/vekjja/ponder.git
cd ponder

# Build
go build -o ponder

# Run
./ponder
```

### Building Docker Image
```bash
docker build -t ponder:local .
```

---

## 🔧 Configuration Options

### OpenAI Settings
- `openAI_endpoint` - API endpoint (default: "https://api.openai.com/v1/")
- `openAI_chat_model` - Chat model (default: "gpt-4")
- `openAI_chat_systemMessage` - System prompt for chat
- `openAI_temperature` - Response randomness (0-2)
- `openAI_maxTokens` - Max response length
- `openAI_topP` - Nucleus sampling parameter
- `openAI_frequencyPenalty` - Repetition penalty
- `openAI_presencePenalty` - Topic diversity penalty

### Image Generation Settings
- `openAI_image_model` - Image model (default: "dall-e-3")
- `openAI_image_size` - Image dimensions (default: "1024x1024")
- `openAI_image_downloadPath` - Save location (default: "~/Ponder/Images/")

### TTS Settings
- `openAI_tts_model` - TTS model (default: "tts-1")
- `openAI_tts_voice` - Default voice (default: "onyx")
- `openAI_tts_speed` - Speech speed (default: "1")
- `openAI_tts_responseFormat` - Audio format (default: "mp3")

### Discord Settings
- `discord_message_context_count` - Messages to include in context
- `discord_bot_systemMessage` - System prompt for Discord bot

---

## 📝 Examples

### Creative Writing Assistant
```bash
ponder "Write a short story about a time-traveling detective"
```

### Code Explanation
```bash
ponder "Explain how blockchain works in simple terms"
```

### Conversational Session
```bash
ponder chat
# Then interact naturally:
You: What's the weather like on Mars?
Ponder: Mars has a thin atmosphere...
You: Could humans survive there?
Ponder: Humans would need...
```

### Generate Art with Download
```bash
ponder image "cyberpunk street scene with neon lights" --download --open
```

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## 📄 License

See [LICENSE](../LICENSE) file for details.

---

## 🔗 Links

- **GitHub Repository**: [https://github.com/seemywingz/ponder](https://github.com/seemywingz/ponder)
- **OpenAI API Documentation**: [https://platform.openai.com/docs](https://platform.openai.com/docs)
- **Discord Developer Portal**: [https://discord.com/developers](https://discord.com/developers)

---

## ⚡ Version

Current Version: **v0.4.3**

---

*Built with ❤️ using Go, OpenAI API, and Bubble Tea*
