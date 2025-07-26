# gator
A CLI blog aggregator that allows multiple users to add, track and follow RSS feeds.

---

## Requirements
You'll need [Postgres](https://www.postgresql.org/) and [Go](https://go.dev/) installed on your computer in order to compile and run this program

---

## Installation

In your command line type the following:
`go install github.com/thmastin/gator@latest`

## Configuration

Before running gator, you need to create a config file named .gatorconfig.json in your home directory. Here’s what it should look like:
...
{
  "db_url": "url for your postgres DB goes here",
  "current_user_name": ""
}
...

- db_url: The full connection string to your Postgres database
- current_user_name: Your username for the application

You can copy this example and change the values for your setup.
Be sure not to share your real .gatorconfig.json—keep it out of version control.
