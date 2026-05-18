![LLMBridge](assets/banner.png)

# LLMBridge

LLMBridge is a concept of a single platform where users can access and switch between any AI provider — cloud or self-hosted — without changing their application code. Just update the config and run one command.

It is built on top of [Bifrost](https://github.com/maximhq/bifrost), an open-source high-performance AI gateway that handles provider routing, failover, and load balancing out of the box. Bifrost turns what would normally require separate integrations for each provider into a single, ready-made solution.

Token usage is metered via [Lago](https://github.com/getlago/lago) — every request is tracked and stored as a usage event, laying the foundation for per-user billing with local payment methods and removing the dependency on international cards.

> **POC note:** events are sent to Lago under a hardcoded `demo-user` customer who does not exist as a registered customer in Lago. Events are stored and the metering layer is working, but they won't be aggregated into invoices until a customer and subscription are set up. For a full application this would be replaced by a real user service with authentication, where each user maps to a Lago customer and is subscribed to a billing plan.

---

## How it works

A chat request flows through the full stack in sequence:

```
User (browser)
    │
    ▼
LLMBridge  :9000        ← Go app — chat UI, token cost, wallet debit
    │
    ▼
Bifrost    :8080        ← AI gateway — routes to the configured provider
    │
    ▼
AI Provider             ← Gemini / Claude / OpenAI / Ollama (local)
    │
    ▼ (response + token count)
LLMBridge
    │
    ├──▶ Wallet debit (in-memory, shown in UI)
    │
    └──▶ Lago API :3000  ← usage event fired async (tokens_used metric)
```

All services run inside the same Docker network — nothing is exposed beyond what's needed.

---

## Services

| Service | Address | Description |
|---|---|---|
| LLMBridge | http://localhost:9000 | Chat UI and API |
| Bifrost | http://localhost:8080 | AI gateway dashboard |
| Lago dashboard | http://localhost:8081 | Billing UI — login: `admin@llmbridge.local` / `password` |
| Lago API | http://localhost:3000 | Billing REST API |
| Ollama | http://localhost:11434 | Local model runtime (only when `MODEL=ollama/*`) |

---

## Lago billing

Lago runs as part of the same `docker compose` stack. On first boot it auto-creates an org, a billable metric (`tokens_used`), and a pre-set API key — no manual setup needed.

After every chat request, LLMBridge fires a usage event to Lago:

```json
{
  "code": "tokens_used",
  "external_customer_id": "demo-user",
  "properties": {
    "input_tokens": 5,
    "output_tokens": 35,
    "total_tokens": 40,
    "model": "gemini/gemini-2.5-flash"
  }
}
```

To verify events are flowing:

```bash
curl http://localhost:3000/api/v1/events \
  -H "Authorization: Bearer llmbridge-lago-dev-key"
```

To explore usage in the dashboard, open **http://localhost:8081** and go to **Events**.

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
