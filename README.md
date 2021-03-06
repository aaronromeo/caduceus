# Caduceus

A tool to tame Gmail

## Getting started

### Required files
* data/credentials.json
  * There is blurb about it [here](https://developers.google.com/workspace/guides/create-credentials#desktop-app)

### Required OAuth scope
While testing I have the following scope enabled
* https://www.googleapis.com/auth/gmail.modify _(which is full access to gmail)_

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
  $ go run main.go

  A tool to help manage Gmail labels, filters and messages.

  Usage:
    caduceus [command]

  Available Commands:
    completion  Generate the autocompletion script for the specified shell
    doctor      A brief description of your command
    fetch       Fetch Gmail resources
    help        Help about any command
    migrate     Run a pending migration file located in the 'migrations' folder

  Flags:
    -h, --help     help for caduceus
    -t, --toggle   Help message for toggle

  Use "caduceus [command] --help" for more information about a command.

  ```

The first time you run the sample, it prompts you to authorize access:
1. Browse to the provided URL in your web browser.
  1. If you're not already signed in to your Google account, you're prompted to sign in. If you're signed in to multiple Google accounts, you are asked to select one account to use for authorization.
1. Click the Accept button.
1. Copy the code you're given, paste it into the command-line prompt, and press Enter.

### Todo
1. HTML email digest with unsubscribe suggestions
1. Use the existing filters to archive contents of the inbox
1. Figure out how to deal with the weekly expiring token

### Background
The starting point for this code and run instructions is from the [Go quickstart](https://developers.google.com/gmail/api/quickstart/go)

Finding data can be done with `jq`
* `cat data/filters.json | jq '.[] | select(.action.addLabelIds | index("Label_33")) | .id'`
* `cat data/labels.json | jq '.[] | select(.name | test("Promo/Webservice/")) | .id'`
* ``for label in `cat data/labels.json | jq '.[] | select(.name | test("Promo/Webservice/")) | .id'`; do cat data/filters.json | jq ".[] | select(.action.addLabelIds | index($label)) | .id"; done | sed -e 's/$/,/'``

### Resources
* [Google Cloud Platform](https://console.cloud.google.com/home/dashboard)
