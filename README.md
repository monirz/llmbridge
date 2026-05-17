# LLMBridge

A proof of concept AI gateway that lets users access any AI provider and switch between them with a single config change — including self-hosted models. Cloud or local, the interface stays the same.

Includes token metering and billing, with support for local payment methods as an alternative to international cards. Full billing integration via Lago is planned but not yet implemented in this POC.

Built on top of [Bifrost](https://github.com/maximhq/bifrost) — an open-source, high-performance AI gateway that unifies 23+ providers behind a single OpenAI-compatible API with ~11 µs overhead at 5,000 RPS.

---



## Switch providers

Create a `.env` file in the project root and set your API key and model:

```env
# Gemini
GEMINI_API_KEY=your_key_here
MODEL=gemini/gemini-2.5-flash

# Claude
ANTHROPIC_API_KEY=your_key_here
MODEL=anthropic/claude-haiku-3-5

# OpenAI
OPENAI_API_KEY=your_key_here
MODEL=openai/gpt-4o-mini

# Ollama (self-hosted, no key needed)
MODEL=ollama/qwen2.5:3b
```

## Run

```bash
git clone https://github.com/monirz/llmbridge.git
cd llmbridge
docker compose up --build
```

Open **http://localhost:9000** in your browser.

---