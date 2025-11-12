# Additional Features for the Forum Project

## forum-security

### Objectives

You must follow the same [principles](../README.md) as the first subject.

For this project you must take into account the security of your forum.

- You should implement a Hypertext Transfer Protocol Secure ([HTTPS](https://developer.mozilla.org/en-US/docs/Glossary/HTTPS)) protocol :

  - Encrypted connection : for this you will have to generate an SSL certificate, you can think of this like a identity card for your website. You can create your certificates or use "Certificate Authorities"(CA's)

  - We recommend you to take a look into [cipher suites](https://en.wikipedia.org/wiki/Cipher_suite).

- The implementation of [Rate Limiting](https://en.wikipedia.org/wiki/Rate_limiting) must be present on this project

- You should encrypt at least the clients passwords. As a Bonus you can also encrypt the database, for this you will have to create a password for your database.

[Sessions](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html#session-management-waf-protections) and [cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies) were implemented in the [previous project](../README.md) but not under-pressure (tested in an attack environment). So this time you must take this into account.

- Clients session cookies should be unique. For instance, the session state is stored on the server and the session should present an unique identifier. This way the client has no direct access to it. Therefore, there is no way for attackers to read or tamper with session state.

### Hints

- You can take a look at the `openssl` manual.
- For the session cookies you can take a look at the [Universal Unique Identifier (UUID)](https://en.wikipedia.org/wiki/Universally_unique_identifier)

### Instructions

- You must handle website errors, HTTPS status.
- You must handle all sort of technical errors.
- The code must respect the [**good practices**](../../good-practices/README.md).
- It is recommended to have **test files** for [unit testing](https://go.dev/doc/tutorial/add-a-test).

### Allowed packages

- All [standard Go](https://golang.org/pkg/) packages are allowed.
- [autocert](https://pkg.go.dev/golang.org/x/crypto/acme/autocert)
- [sqlite3](https://github.com/mattn/go-sqlite3)
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [gofrs/uuid](https://github.com/gofrs/uuid) or [google/uuid](https://github.com/google/uuid)

This project will help you learn about :

- HTTPS
- [Cipher suites](https://en.wikipedia.org/wiki/Cipher_suite)
- Goroutines
- Channels
- Rate Limiting
- Encryption
  - password
  - session/cookies
  - Universal Unique Identifier (UUID)

---

## forum-moderation

### Objectives

You must follow the same [principles](../README.md) as the first subject.

The `forum moderation` will be based on a moderation system. It must present a moderator that, depending on the access level of a user or the forum set-up, approves posted messages before they become publicly visible.

- The filtering can be done depending on the categories of the post being sorted by irrelevant, obscene, illegal or insulting.

For this optional you should take into account all types of users that can exist in a forum and their levels.

You should implement at least 4 types of users :

#### Guests

- These are unregistered-users that can neither post, comment, like or dislike a post. They only have the permission to **see** those posts, comments, likes or dislikes.

#### Users

- These are the users that will be able to create, comment, like or dislike posts.

#### Moderators

- Moderators, as explained above, are users that have a granted access to special functions :
  - They should be able to monitor the content in the forum by deleting or reporting post to the admin
- To create a moderator the user should request an admin for that role

#### Administrators

- Users that manage the technical details required for running the forum. This user must be able to :
  - Promote or demote a normal user to, or from a moderator user.
  - Receive reports from moderators. If the admin receives a report from a moderator, he can respond to that report
  - Delete posts and comments
  - Manage the categories, by being able to create and delete them.

### Instructions

- You must handle website errors, HTTPS status.
- You must handle all sort of technical errors.
- The code must respect the [**good practices**](../../good-practices/README.md).
- It is recommended to have **test files** for [unit testing](https://go.dev/doc/tutorial/add-a-test).

### Allowed packages

- All [standard go](https://golang.org/pkg/) packages are allowed.
- [sqlite3](https://github.com/mattn/go-sqlite3)
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [gofrs/uuid](https://github.com/gofrs/uuid) or [google/uuid](https://github.com/google/uuid)

This project will help you learn about:

- Moderation System
- User access levels

---

## forum-image-upload

### Objectives

You must follow the same [principles](../README.md) as the first subject.

In `forum image upload`, registered users have the possibility to create a post containing an image as well as text.

- When viewing the post, users and guests should see the image associated to it.

There are several extensions for images like: JPEG, SVG, PNG, GIF, etc. In this project you have to handle at least JPEG, PNG and GIF types.

The max size of the images to load should be 20 mb. If there is an attempt to load an image greater than 20mb, an error message should inform the user that the image is too big.

### Hints

- Be cautious with the size of the images.

### Instructions

- The backend must be written in **Go**.
- You must handle website errors.
- The code must respect the [good practices](../../good-practices/README.md)
- It is recommended to have **test files** for [unit testing](https://go.dev/doc/tutorial/add-a-test).

### Allowed packages

- All [standard go](https://golang.org/pkg/) packages are allowed.
- [sqlite3](https://github.com/mattn/go-sqlite3)
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [gofrs/uuid](https://github.com/gofrs/uuid) or [google/uuid](https://github.com/google/uuid)

This project will help you learn about:

- Image manipulation
- Image types

---

## authentication

### Objectives

The goal of this project is to implement, into your forum, new ways of authentication. You have to be able to register and to login using at least Google and Github authentication tools.

Some examples of authentication means are:

- Facebook
- GitHub
- Google

### Instructions

- Your project must have implemented at least the two authentication examples given.
- Your project must be written in **Go**.
- The code must respect the [**good practices**](../../good-practices/README.md).

### Allowed packages

- All [standard go](https://golang.org/pkg/) packages are allowed.
- [sqlite3](https://github.com/mattn/go-sqlite3)
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [gofrs/uuid](https://github.com/gofrs/uuid) or [google/uuid](https://github.com/google/uuid)

This project will help you learn about:

- [Sessions](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html#session-management-waf-protections) and [cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies)
- Protecting routes

---

## forum-advanced-features

### Objectives

You must follow the same [principles](../README.md) as the first subject.

In `forum advanced features`, you will have to implement the following features :

- You will have to create a way to notify users when their posts are :

  - liked/disliked
  - commented

- You have to create an activity page that tracks the user own activity. In other words, a page that :

  - Shows the user created posts
  - Shows where the user left a like or a dislike
  - Shows where and what the user has been commenting. For this, the comment will have to be shown, as well as the post commented

- You have to create a section where you will be able to Edit/Remove posts and comments.

We encourage you to add any other additional features that you find relevant.

### Instructions

- The backend must be written in **Go**
- The code must respect the [good practices](../../good-practices/README.md)
- It is recommended to have **test files** for [unit testing](https://go.dev/doc/tutorial/add-a-test).

### Allowed packages

- All [standard go](https://golang.org/pkg/) packages are allowed.
- [sqlite3](https://github.com/mattn/go-sqlite3)
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [gofrs/uuid](https://github.com/gofrs/uuid) or [google/uuid](https://github.com/google/uuid)

### This project will help you learn about:

- Real-time notifications
- Users activity tracking