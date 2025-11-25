# AI Service Microservice

[![CI Status](https://github.com/Triune-Oracle/Logos_Agency/actions/workflows/ci.yml/badge.svg)](https://github.com/Triune-Oracle/Logos_Agency/actions/workflows/ci.yml)
[![Code Coverage](https://img.shields.io/badge/coverage-0%25-red)](https://github.com/Triune-Oracle/Logos_Agency)

## 📚 Documentation & Guides

📖 [Getting Started Guide](docs/GETTING_STARTED.md)
*   [Architecture](docs/ARCHITECTURE.md)
*   [Performance](docs/PERFORMANCE.md)
*   [Roadmap](docs/ROADMAP.md)
*   [Monetization](docs/MONETIZATION.md)

## Overview
This microservice provides advanced conversation analysis leveraging OpenAI APIs. Supports sentiment, archetype, emotion, theme, and communication pattern analysis.

## Installation

```bash
git clone <repo-url>
cd ai-service
npm install
cp .env.example .env
# Edit .env with your API keys
npm start
```

## API Endpoints

| Endpoint            | Method | Description                      |
|---------------------|--------|--------------------------------|
| `/analyze/sentiment`| POST   | Analyze conversation sentiment   |
| `/analyze/archetype`| POST   | Detect archetypes in text        |
| `/analyze/emotion`  | POST   | Detailed emotional profile       |
| `/analyze/summary`  | POST   | Generate conversation summary    |
| `/analyze/insights` | POST   | Generate actionable insights     |
| `/health`           | GET    | Service health status            |

## Prompts

Stored in `src/prompts/` for fine-tuning.

## Testing

Run `npm test` to execute unit and integration tests.

## Contribution

Pull requests welcome. Follow code style, write tests, update docs.

## Deployment

Supports containerized Docker or Kubernetes deployment. See `.github/workflows` for CI/CD.

## License

MIT License
