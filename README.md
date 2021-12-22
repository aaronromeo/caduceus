# Caduceus

A tool to help me manage my mailbox

## Getting started

### Required files
* credentials.json
  * There is blurb about it [here](https://developers.google.com/workspace/guides/create-credentials#desktop-app)

### Required OAuth scope
While testing I have the following scopes enabled
* https://www.googleapis.com/auth/gmail.labels _(which I think is the needed one)_
* XXX

### Running the code
#### Prepare the workspace
* Set the GOPATH environment variable to your working directory.
* Get the Gmail API Go client library and OAuth2 package using the following commands:
  ```bash
  go get -u google.golang.org/api/gmail/v1
  go get -u golang.org/x/oauth2/google
  ```

#### Run the code
Build and run the sample using the following command from your working directory:
  ```bash
  go run quickstart.go
  ```

The first time you run the sample, it prompts you to authorize access:
1. Browse to the provided URL in your web browser.
  1. If you're not already signed in to your Google account, you're prompted to sign in. If you're signed in to multiple Google accounts, you are asked to select one account to use for authorization.
1. Click the Accept button.
1. Copy the code you're given, paste it into the command-line prompt, and press Enter.


### Background
The starting point for this code and run instructions is from the [Go quickstart](https://developers.google.com/gmail/api/quickstart/go)


### Resources
* [Google Cloud Platform](https://console.cloud.google.com/home/dashboard)
