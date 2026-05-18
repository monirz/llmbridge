# LLMBridge

LLMBridge is a concept of a single platform where users can access and switch between any AI provider — cloud or self-hosted — without changing their application code. Just update the config and run one command.

It is built on top of [Bifrost](https://github.com/maximhq/bifrost), an open-source high-performance AI gateway that handles provider routing, failover, and load balancing out of the box. Bifrost turns what would normally require separate integrations for each provider into a single, ready-made solution.

Token usage is metered via [Lago](https://github.com/getlago/lago) — every request is tracked and stored as a usage event, laying the foundation for per-user billing with local payment methods and removing the dependency on international cards.

> **POC note:** events are sent to Lago under a hardcoded `demo-user` customer who does not exist as a registered customer in Lago. Events are stored and the metering layer is working, but they won't be aggregated into invoices until a customer and subscription are set up. For a full application this would be replaced by a real user service with authentication, where each user maps to a Lago customer and is subscribed to a billing plan.

---



## Run

Clone the repo and cd into it:

```bash
git clone https://github.com/monirz/llmbridge.git
cd llmbridge
```

Then start all services:

```bash
docker compose up --build
```

Once everything is up:

- **http://localhost:9000** — LLMBridge chat UI
- **http://localhost:8080** — Bifrost gateway dashboard
- **http://localhost:8081** — Lago billing dashboard
- **http://localhost:3000** — Lago API

By default it runs with the self-hosted `qwen2.5:3b` model via Ollama — no API key needed. Bifrost, Ollama, Lago, and LLMBridge are all managed under the same `docker compose`.

> Lago auto-creates an org on first boot. The API key is pre-set to `llmbridge-lago-dev-key` — no manual step needed.

---

## Switch providers

Create a `.env` file in the project root. See `.env.example` for all available variables.

**Ollama — local, no API key (default):**

```env
MODEL=ollama/qwen2.5:3b
```

**Gemini:**

```env
GEMINI_API_KEY=your_key_here
MODEL=gemini/gemini-2.5-flash
```

**Claude:**

```env
ANTHROPIC_API_KEY=your_key_here
MODEL=anthropic/claude-haiku-3-5
```

**OpenAI:**

```env
OPENAI_API_KEY=your_key_here
MODEL=openai/gpt-4o-mini
```

Ollama starts automatically when `MODEL` is set to `ollama/*` or not set at all. Any other model skips Ollama and the download entirely.

---