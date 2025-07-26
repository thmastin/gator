# gator
A CLI blog aggregator that allows multiple users to add, track and follow RSS feeds.

---

## Requirements
You'll need [Postgres](https://www.postgresql.org/) and [Go](https://go.dev/) installed on your computer in order to compile and run this program

---

## Installation

First, install `gator` with:
`go install github.com/thmastin/gator@latest`


This command downloads, builds, and installs `gator` to your Go bin directory. By default, this is `$HOME/go/bin` on Linux and macOS, or `%USERPROFILE%\go\bin` on Windows.

**Make sure your Go bin directory is in your `PATH`** so you can run `gator` from anywhere.

On Linux/macOS, you can add this line to your `.bashrc`, `.zshrc`, or `.profile`:

```sh
export PATH="$PATH:$HOME/go/bin"
```
On Windows, add `%USERPROFILE%\go\bin` to your system’s PATH environment variable.
[See here for instructions on editing your PATH on Windows.](https://stackoverflow.com/questions/1618280/where-can-i-set-path-to-make-exe-on-windows)
---

## Configuration

Before running `gator`, create a config file named `.gatorconfig.json` in your home directory. Here’s what it should look like:

- Linux/macOS: `/home/yourname/.gatorconfig.json`
- Windows: `C:\Users\YourName\.gatorconfig.json`

```json
{
  "db_url": "url for your postgres DB goes here",
  "current_user_name": ""
}
```

- db_url: The full connection string to your Postgres database
- current_user_name: Your username for the application

You can copy this example and change the values for your setup.
Be sure not to share your real .gatorconfig.json—keep it out of version control.

---
## Commands

All commands use the following format: `gator <command> <arguments>`

Examples:
- `register <username>`: Registers a user and logs them in.
- `users`: Shows all users in the database.
- `login <username>`: Changes active user to the provided username.
- `addfeed <name> <feed url>`: Adds an RSS feed and follows it.
- `follow <feed url>`: Follows a feed another user has added.
- `feeds`: Displays all feeds.
- `following`: Shows feeds the logged in user is following.
- `unfollow <feed url>`: Unfollows a feed.
- `agg <time>`: Starts the aggregator to fetch the oldest feeds the user is following. The time argument sets how often to fetch, in hour-minute-second format (e.g., `5m20s` will pull feeds every 5 minutes and 20 seconds.)
- `browse <number of posts>`: Shows the most recent posts from followed feeds. Defaults to 2, but can be set by the user.