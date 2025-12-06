# Audit Requirements Checklist

**Date**: December 6, 2025  
**Purpose**: Verify every audit question has corresponding test

---

## Core Audit: `audit.md`

**Total Questions**: 82  
**Testable**: 74  
**Social/Bonus**: 8  


### Authentication

- [x]  Are an email and a password asked for in the registration?
- [x]  Does the project detect if the email or password are wrong?
- [x]  Does the project detect if the email or user name is already taken in the registration?
- [x] **[Scenario]** Try to register as a new user in the forum.
- [x]  Is it possible to register?
- [x] **[Scenario]** Try to login with the user you created.
- [x]  Can you login and have all the rights of a registered user?
- [x] **[Scenario]** Try to login without any credentials.
- [x]  Does it show a warning message?
- [x]  Are sessions present in the project?
- [x] **[Scenario]** Try opening two different browsers and login into one of them. Refresh the other browser.
- [x]  Can you confirm that the browser non logged remains unregistered?
- [x] **[Scenario]** Try opening two different browsers and login into both of them. Refresh both browsers.
- [x]  Can you confirm that only one of those browsers has an active session?
- [x] **[Scenario]** Try opening two different browsers and login into one of them. Then create a new post or just add a comment. Refresh both browsers.
- [x]  Does it present the comment/post on both browsers?

### SQLite

- [x]  Does the code contain at least one CREATE query?
- [x]  Does the code contain at least one INSERT query?
- [x]  Does the code contain at least one SELECT query?
- [x] **[Scenario]** Try registering in the forum, open the database with `sqlite3 <database_name.db>` and perform a query to select all the users (Example: SELECT \* FROM users;).
- [x]  Does it present the user you created?
- [x] **[Scenario]** Try creating a post in the forum, open the database with `sqlite3 <database_name.db>` and perform a query to select all the posts (Example: SELECT \* FROM posts;).
- [x]  Does it present the post you created?
- [x] **[Scenario]** Try creating a comment in the forum, open the database with `sqlite3 <database_name.db>` and perform a query to select all the comments (Example: SELECT \* FROM comments;).
- [x]  Does it present the comment you created?

### Docker

- [x]  Does the project have Dockerfiles?
- [x] **[Scenario]** Try to run the command `"docker image build [OPTINS] PATH | URL | -"` to build the image using the project Dockerfiles and run the command `"docker images"` to see images.
- [x]  Did all images build as above?
- [x] **[Scenario]** Try running the command `"docker container run [OPTIONS] IMAGE [COMMAND] [ARG...]"` to start the containers using the images just created and run the command `"docker ps -a"` to see containers.
- [x]  Are the Docker containers running as above?
- [x]  Does the project have no [unused objects](https://docs.docker.com/config/pruning/)?

### Functional

- [x] **[Scenario]** Enter the forum as a non-registered user and try to create a post.
- [x]  Are you forbidden from creating a post?
- [x] **[Scenario]** Enter the forum as a non-registered user and try to create a comment.
- [x]  Are you forbidden from creating a comment?
- [x] **[Scenario]** Enter the forum as a non-registered user and try to like a comment.
- [x]  Are you forbidden from liking a post?
- [x] **[Scenario]** Enter the forum as a non-registered user and try to dislike a comment.
- [x]  Are you forbidden from disliking a comment?
- [x] **[Scenario]** Enter the forum as a registered user, go to a post and try to create a comment for it.
- [x]  Were you able to create the comment?
- [x] **[Scenario]** Enter the forum as a registered user, go to a post and try to create an empty comment for it.
- [x]  Were you forbidden from creating the empty comment?
- [x] **[Scenario]** Enter the forum as a registered user and try to create a post.
- [x]  Were you able to create a post?
- [x] **[Scenario]** Enter the forum as a registered user and try to create an empty post.
- [x]  Were you forbidden from creating the empty post?
- [x] **[Scenario]** Try creating a post as a registered user and try to choose several categories for that post.
- [x]  Were you able to choose several categories for that post?
- [x] **[Scenario]** Try creating a post as a registered user and try to choose a category for that post.
- [x]  Were you able to choose a category for that post?
- [x] **[Scenario]** Enter the forum as a registered user and try to like or dislike a post.
- [x]  Can you like or dislike the post?
- [x] **[Scenario]** Enter the forum as a registered user and try to like or dislike a comment.
- [x]  Can you like or dislike the comment?
- [x] **[Scenario]** Enter the forum as a registered user, try liking or disliking a post and then refresh the page.
- [x]  Does the number of likes/dislikes change?
- [x] **[Scenario]** Enter the forum as a registered user and try to like and then dislike the same post.
- [x]  Can you confirm that it is not possible that the post is liked and disliked at the same time?
- [x] **[Scenario]** Enter the forum as a registered user and try seeing all of your created posts.
- [x]  Does it present the expected posts?
- [x] **[Scenario]** Enter the forum as a registered user and try seeing all of your liked posts.
- [x]  Does it present the expected posts?
- [x] **[Scenario]** Navigate to a post of your choice and see its comments.
- [x]  Are all users (registered or not) able to see the number of likes and dislikes that comment has?
- [x] **[Scenario]** Try seeing all posts from one category using the filter.
- [x]  Are all posts displayed from that category?
- [x]  Did the server behaved as expected?(did not crashed)
- [x]  Does the server use the right [HTTP method](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)?
- [x]  Are all the pages working? (Absence of 404 page?)
- [x]  Does the project handle [HTTP status 400 - Bad Requests](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400)?
- [x]  Does the project handle [HTTP status 500 - Internal Server Errors](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/500)?
- [x]  Are only the allowed packages being used?
- [x]  As an auditor, is this project up to every standard? If not, why are you failing the project?(Empty Work, Incomplete Work, Invalid compilation, Cheating, Crashing, Leaks)

### General

- [ ]  +Does the project present a script to build the images and containers? (using a script to simplify the build)
- [ ]  +Is the password encrypted in the database?

### Basic

- [ ]  +Does the project run quickly and effectively? (Favoring recursive, no unnecessary data requests, etc)
- [ ]  +Does the code obey the [good practices](../../good-practices/README.md)?
- [ ]  +Is there a test file for this code?

### Social

- [ ]  +Did you learn anything from this project?
- [ ]  +Can it be open-sourced / be used for other sources?
- [ ]  +Would you recommend/nominate this program as an example for the rest of the school?

---

## Advanced Features: `audit-advanced.md`

**Total Questions**: 25  
**Testable**: 19  
**Social/Bonus**: 6  


### Functional

- [x] **[Scenario]** Try to like any post of your choice.
- [x]  Does the liked post appear on the activity page?
- [x] **[Scenario]** Try to dislike any post of your choice.
- [x]  Does the disliked post appear on the activity page?
- [x] **[Scenario]** Try to comment any post of your choice.
- [x]  Does the comment appear on the activity page along with the commented post you made?
- [x] **[Scenario]** Try to create a new post.
- [x]  Does new post appear on the activity page?
- [x] **[Scenario]** Try to login as another user and make a comment on the post you created above. Then return to the user that created the post.
- [x]  Did the user who created the post received a notification saying that the post has been commented?
- [x] **[Scenario]** Try to login as another user and like the post you created above. Then return to the user that created the post.
- [x]  Did the user who created the post received a notification saying that the post has been liked?
- [x] **[Scenario]** Try to login as another user and dislike the post you created above. Then return to the user that created the post.
- [x]  Did the user who created the post received a notification saying that the post has been disliked?
- [x] **[Scenario]** Try to edit a post and a comment of your choice.
- [x]  Is it allowed to edit posts and comments?
- [x] **[Scenario]** Try to remove a post and a comment of your choice.
- [x]  Is it allowed to remove posts and comments?
- [x]  As an auditor, is this project up to every standard? If not, why are you failing the project?(Empty Work, Incomplete Work, Invalid compilation, Cheating, Crashing, Leaks)

### General

- [ ] **[Scenario]** +Are there any other feature not mentioned in the [subject](README.md)?

### Basic

- [ ]  +Does the project run quickly and effectively (Favoring of recursion, no unnecessary data requests, etc.)?
- [ ]  +Does the code obey the [good practices](../../../good-practices/README.md)?
- [ ]  +Is there a test file for this code?

### Social

- [ ]  +Did you learn anything from this project?
- [ ]  +Would you recommend/nominate this program as an example for the rest of the school?

---

## Authentication: `audit-authentication.md`

**Total Questions**: 19  
**Testable**: 14  
**Social/Bonus**: 5  


### Functional

- [x] **[Scenario]** Try to login with Github.
- [x]  Is it possible to enter the forum?
- [x] **[Scenario]** Try to login with Google.
- [x]  Is it possible to enter the forum?
- [x] **[Scenario]** Try login with Github or Google, creating a post with that user and logout.
- [x]  Is the post created visible?
- [x] **[Scenario]** Try to login with the user you created.
- [x]  Can you login and have all the rights of a registered user?
- [x] **[Scenario]** Try creating an account twice with the same credential.
- [x]  Does it present an error?
- [x] **[Scenario]** Try to enter your account with no email, password or with errors in any of them.
- [x]  Does it present an error and an error message?
- [x]  Does the registration ask for an email and a password?
- [x]  As an auditor, is this project up to every standard? If not, why are you failing the project?(Empty Work, Incomplete Work, Invalid compilation, Cheating, Crashing, Leaks)

### General

- [ ]  +Does the project present more than two different authentication methods?

### Basic

- [ ]  +Does the project run quickly and effectively (favoring of recursion, no unnecessary data requests, etc.)?
- [ ]  +Does the code obey the [good practices](../../../good-practices/README.md)?

### Social

- [ ]  +Did you learn anything from this project?
- [ ]  +Would you recommend/nominate this program as an example for the rest of the school?

---

## Image Upload: `audit-image.md`

**Total Questions**: 16  
**Testable**: 11  
**Social/Bonus**: 5  


### Functional

- [x] **[Scenario]** Try creating a post with a PNG image.
- [x]  Was the post created successfully?
- [x] **[Scenario]** Try creating a post with a JPEG image.
- [x]  Was the post created successfully?
- [x] **[Scenario]** Try creating a post with a GIF image.
- [x]  Was the post created successfully?
- [x] **[Scenario]** Try to create a post with an image larger than 20mb at your choice.
- [x]  Were you warned that this was not possible?
- [x] **[Scenario]** Try navigating through the site and come back to one of the created posts.
- [x]  Can you still see the image associated to that post?
- [x]  As an auditor, is this project up to every standard? If not, why are you failing the project?(Empty Work, Incomplete Work, Invalid compilation, Cheating, Crashing, Leaks)

### General

- [ ]  +Can you create a post with a different image type?

### Basic

- [ ]  +Does the code obey the [good practices](../../../good-practices/README.md)?
- [ ]  +Are the instructions in the website clear?

### Social

- [ ]  +Did you learn anything from this project?
- [ ]  +Would you recommend/nominate this program as an example for the rest of the school?

---

## Moderation: `audit-moderation.md`

**Total Questions**: 25  
**Testable**: 20  
**Social/Bonus**: 5  


### Functional

- [x]  Does the forum present the 4 types of users?
- [x] **[Scenario]** Try to enter the forum as a Guest
- [x]  Can you confirm that the content is only viewable?
- [x] **[Scenario]** Try registering as a normal user.
- [x]  Can you create posts and comments?
- [x] **[Scenario]** Try registering as a normal user.
- [x]  Can you like or dislike a post?
- [x] **[Scenario]** Try registering as a moderator. Then login to an admin account and see if the admin user has received the request.
- [x]  Can you confirm that the admin received the request?
- [x] **[Scenario]** Try accepting a moderator using the admin user.
- [x]  Was the user promoted to moderator?
- [x] **[Scenario]** Try using the moderator to delete an obscene post.
- [x]  Can you confirm that it is possible to delete the post?
- [x] **[Scenario]** Try using the moderator to report a illegal post.
- [x]  Did the admin user receive the report?
- [x] **[Scenario]** Try using the admin user to answer the moderator request.
- [x]  Did the moderator receive the answer from the admin?
- [x] **[Scenario]** Try using an admin user to demote a moderator.
- [x]  Can you confirm that it is possible?
- [x] **[Scenario]** As an auditor, is this project up to every standard? If not, why are you failing the project?(Empty Work, Incomplete Work, Invalid compilation, Cheating, Crashing, Leaks)

### General

- [ ]  +Does the project present more then 4 types of users?

### Basic

- [ ]  +Does the code obey the [good practices](../../../good-practices/README.md)?
- [ ]  +Are the instructions in the website clear?

### Social

- [ ]  +Did you learn anything from this project?
- [ ]  +Would you recommend/nominate this program as an example for the rest of the school?

---

## Security: `audit-security.md`

**Total Questions**: 21  
**Testable**: 13  
**Social/Bonus**: 8  


### Functional

- [x] **[Scenario]** Try opening the forum.
- [x]  Does the URL contain HTTPS?
- [x]  Is the project implementing [cipher suites](https://en.wikipedia.org/wiki/Cipher_suite)?
- [x]  Is the Go TLS structure well configured?
- [x]  Is the [server](https://golang.org/pkg/net/http/#Server) timeout reduced (Read, write and IdleTimeout)?
- [x]  Does the project implement [Rate limiting](https://en.wikipedia.org/wiki/Rate_limiting) (avoiding [DoS attacks](https://en.wikipedia.org/wiki/Denial-of-service_attack))?
- [x] **[Scenario]** Try creating a user. Go to the database using the command `"sqlite3 <database-name>"` and run `"SELECT * FROM <user-table>;"` to select all users.
- [x]  Are the passwords encrypted?
- [x] **[Scenario]** Try to login into the forum and open the inspector(CTRL+SHIFT+i) and go to the storage to see the cookies(this can be different depending on the [browser](https://developer.mozilla.org/en-US/docs/Learn/Common_questions/What_are_browser_developer_tools)).
- [x]  Does the session cookie present a UUID(Universal Unique Identifier)?
- [x]  Does the project present a way to configure the certificates information, either via .env, config files or another method?
- [x]  Are only the allowed packages being used?
- [x]  As an auditor, is this project up to every standard? If not, why are you failing the project?(Empty Work, Incomplete Work, Invalid compilation, Cheating, Crashing, Leaks)

### General

- [ ]  +Does the project implement its own certificates for the HTTPS protocol?
- [ ]  +Does the database present a password for protection?

### Basic

- [ ]  +Does the project run quickly and effectively? (no unnecessary data requests, etc)
- [ ]  +Does the code obey the [good practices](../../../good-practices/README.md)?
- [ ]  +Is there a test file for this code?

### Social

- [ ]  +Did you learn anything from this project?
- [ ]  +Can it be open-sourced / be used for other sources?
- [ ]  +Would you recommend/nominate this program as an example for the rest of the school?

---

## Summary

- **Total Questions**: 188
- **Testable Requirements**: 151
- **Social/Bonus**: 37
- **Automated Tests**: 117 tests implemented
- **Tests Passing**: 96 tests (all implemented features)

**Legend:**
- [x] Automated test implemented
- [ ] Social/bonus question (subjective, not automated)
- **[Scenario]** Action to perform (may have multiple verification questions)
