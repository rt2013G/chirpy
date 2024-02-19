# chirpy
A web server built in Go working as a Twitter back-end clone

## Endpoints
- /app - GET - Home page

## API
- api/healthz - GET - Checks readiness
- api/reset - GET - Resets hit count
- api/chirps - GET - Gets all the chirps
- api/chirps - POST - Creates a new chirp
- api/chirps{chirpID} - GET - Gets chirp with chirpID
- api/chirps{chirpID} - DELETE - Deletes chirp with chirpID
- api/users - POST - Creates a new user
- api/users - PUT - Updates a user
- api/login - POST - Logins with credentials, return JWT access token
- api/refresh - POST - Refreshes access token
- api/revoke - POST - Revokes token

## Contributing
Clone the repo
```bash
git clone https://github.com/rt2013G/chirpy && cd chirpy
```

Build and run
```bash
go build -o out && ./out --debug
```
