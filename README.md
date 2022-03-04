# Dashy
Create a personalized private dashboard

## Getting Started
1. Fork the repository as you will need to customize it
1. Edit `config/config.yml` with all your dashboard needs
1. Create a `.env` file if you want local env vars easily
1. Install and run redis
1. Run feed: `go run ./feed`
1. Run web: `go run ./web`

## Heroku
1. Create a new heroku app
1. Configure all the environment variables you need for your dashboard
1. Push dashy: `git push heroku main`
1. Add the "Heroku Scheduler" app
1. Create a new job and set the job command to: `feed`

## Configuration
Docs coming soon... very much unstable
