# AI Service Microservice

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
