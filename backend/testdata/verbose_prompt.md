以下の通り、お願いします。Please implement user authentication. Here is the request.

## Goal
Add JWT-based login API with refresh tokens and session management.

## Context
We have an existing Go monolith. Based on the above, here is the current state.
The team discussed this in the last sprint. Please review carefully.
Here is additional background that repeats the same points.
The service runs on Kubernetes. The database is PostgreSQL.
Based on the above, stakeholders want OAuth2 compatibility eventually.

## Rules
- MUST use clean code best practices at all times
- MUST ONLY use approved libraries from the internal catalog
- Follow project conventions documented in CONTRIBUTING.md
- npm install dependencies before starting any work
- git commit frequently with descriptive messages
- ALWAYS write comprehensive unit tests for every function
- ONLY merge when CI is green

## Skills
When implementing auth, read the security playbook first.
Multi-step workflow: design, implement, test, document, deploy.

## History
User: Can we add auth?
Assistant: Sure, what kind?
User: JWT please
User: Also add refresh tokens
User: And social login
User: And admin panel
User: Can we add auth?
Assistant: We already discussed JWT.
User: What about rate limiting?
User: And audit logging?
User: And MFA?
User: And password reset flow?
User: And email verification?
User: And account lockout policy?

```go
func main() {
    fmt.Println("hello")
}
```

```go
func authenticate(token string) error {
    return nil
}
```

Please make it fast and also be very thorough with security review.
The deadline is tight but quality cannot be compromised.
