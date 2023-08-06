# Iris

[![Go](https://github.com/trynoice/iris/actions/workflows/go.yaml/badge.svg)](https://github.com/trynoice/iris/actions/workflows/go.yaml)
[![codecov](https://codecov.io/github/trynoice/iris/branch/main/graph/badge.svg?token=7SA1GWRIJY)](https://codecov.io/github/trynoice/iris)
[![License](https://img.shields.io/github/license/trynoice/iris.svg)](LICENSE)

Iris is a CLI tool for dispatching templated emails.

## Supported Services

- AWS SES
- SMTP

## Install

Download a suitable binary from the [latest GitHub
release](https://github.com/trynoice/iris/releases/latest). Or if you have
`$GOPATH/bin` in your system path, you can use `go install`.

```console
go install github.com/trynoice/iris@latest
```

## Usage

First, create working files using iris init. You can omit the positional
argument to initialise files in the current working directory. These files
contain a sensible configuration, a minimal email template and placeholder data
to get you started quickly.

```console
$ iris init sample-email
creating directory sample-email
creating file sample-email/.iris.yaml
creating file sample-email/recipients.csv
creating file sample-email/subject.txt
creating file sample-email/body.txt
creating file sample-email/body.html
creating file sample-email/default.csv
```

### Working Files

- `.iris.yaml`: Iris configuration for this email template.
- `subject.txt`: A [Go template](https://pkg.go.dev/text/template) containing
  the subject line of the email.
- `body.txt`: A [Go Template](https://pkg.go.dev/text/template) containing the
  body of the email in plain text format.
- `body.html`: A [Go Template](https://pkg.go.dev/text/template) containing the
  body of the email in HTML format.
- `recipients.csv`: Data for rendering the email templates.
- `default.csv`: Optional fallback values for missing values in recipients' data
  file. You can also use it to inject data that remains the same for all
  recipients.

### Configuration

```yaml
service:
    # If using AWS SES backend.
    awsSes:
        # If `true`, automatically load AWS configuration from ~/.aws or env vars.
        useSharedConfig: true
        # AWS region for SES if `useSharedConfig` is `false`.
        region:
        # AWS configuration profile if `useSharedConfig` is `false`.
        profile:

    # If using SMTP server backend.
    smtp:
        # Host name of the smtp server.
        host:
        # Port of the smtp server.
        port:
        # Username for authenticating on the smtp server.
        username:
        # Password for authenticating on the smtp server.
        password:
        # One of 'none', 'ssl', 'tls', 'ssl/tls' (default), 'starttls'.
        encryption:

    # API calls per second.
    rateLimit: 10
    # Number of retries before exiting with error on failing an API call.
    retries: 3
message:
    # An address for the 'From' email header.
    sender: Iris CLI <iris@trynoice.com>
    # A list of addresses for the 'Reply-To' email header.
    replyToAddresses:
        - Noice App <trynoiceapp@gmail.com>
    # Data for rendering the email templates. It must be in the same directory
    # as this configuration.
    recipientDataCsvFile: recipients.csv
    # (Optional) Fallback values for missing values in recipients' data file.
    # You can also use it to inject data that remains the same for all
    # recipients. It must be in the same directory as this configuration. It
    # must contain only two rows: headers and values.
    defaultDataCsvFile: default.csv
    # Name of the column containing emails of recipients in `recipients.csv`.
    recipientEmailColumnName: Email
    # If true, minify rendered HTML before composing the email message.
    minifyHtml: true
```

### Send Emails

Verify rendered email for a template with a dry run.

```console
$ iris send sample-email --dry-run
dispatching to jack@example.test
+-----------+---------------------------------------------------------+
| Subject   | Hello Jack                                              |
+-----------+---------------------------------------------------------+
| Text Body | Iris is a CLI tool for sending templated bulk emails.   |
|           |                                                         |
|           | You can inject data into templates, e.g. a date -       |
|           | January 2006 or your email - jack@example.test.         |
+-----------+---------------------------------------------------------+
| HTML Body | <!doctype html><html><head><meta name=viewport          |
|           | content="width=device-width,initial-scale=1"><meta      |
|           | charset=utf-8><title>Hello                              |
|           | Jack</title></head><body><p>Iris is a CLI tool for      |
|           | sending templated bulk emails.</p><p>You can inject     |
|           | data into templates, e.g. a date - January 2006 or your |
|           | email - jack@example.test.</p></body></html>            |
+-----------+---------------------------------------------------------+
```

And dispatch emails.

```console
$ iris send sample-email
confirm sending emails? [y/n] y
dispatching to jack@example.test
```

## License

[Apache License 2.0](LICENSE)
