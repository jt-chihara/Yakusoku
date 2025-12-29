package main

// This file contains the UserClient that will be tested with contract tests.
// The actual client implementation is in main.go.
// This file exists to show the interface being tested.

/*
UserClient API:

- GetUser(id int) (*User, error)
  Fetches a user by ID from UserService.

  Expected Request:
    GET /users/{id}

  Expected Response (200 OK):
    {
      "id": 1,
      "name": "Alice",
      "email": "alice@example.com"
    }

  Expected Response (404 Not Found):
    {
      "error": "User not found"
    }
*/
