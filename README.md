# `invoice` command line tool

__`invoice`__ lets you quickly and intuitively exchange invoice file(s) with the
Italian Servizio di Interscambio (SDI) from the command line. It leverages the [Invoicetronic API][1].

You don't need to learn the [Invoicetronic API][1] or any programming language to
send and receive invoices (SDKs for common programming languages are also
[available][2]).

## At a glance

Sending an invoice can be as simple as issuing this command:

```bash
invoice send file1.xml
```
Say you have many files to upload, you can use wildcards:

```bash
invoice send *.xml --delete
invoice send file1.xml file2.xml file3.xml
```

In the first line above, we're also deleting files from the disk once
successfully uploaded. 

Receiving files is also super simple:

```bash
invoice receive --unread
```

The above will download all new invoices and store them in the current directory. 

## Installation

See the [Installation guide][3].

## Quickstart

Quickstart and context available at the [Invoicetronic website][1].

[1]: https://invoicetronic.com/docs/quickstart/invoice-quickstart/
[2]: https://invoicetronic.com/docs/sdk/
[3]: https://invoicetronic.com/docs/cli/#installation-guide