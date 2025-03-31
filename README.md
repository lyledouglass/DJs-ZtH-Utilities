# Death Jesters & Zeroes to Heroes Utilities Bot

This repository is the source code for the Death Jesters & Zeroes to Heroes Utilities Bot, a Discord bot, written in Go, designed to help manage and enhance the experience of the Death Jesters and Zeroes to Heroes community members. This guide will provide an comprehensive overview of the bot's features, installation instructions, and usage guidelines

## Building the bot

### Prerequisites

* **Go**: Ensure you have Go installed on your system. You can download it from the official [Go website](https://golang.org/dl/)
* **Discord Bot Token**: You will need to create a Discord bot and obtain a token. Follow the instructions in the [Discord Developer Portal](https://discord.com/developers/applications) to create your bot and get the token.
* **A Linux based system**: This bot was developed and tested on a Linux based system. While it may work on other operating systems, it is recommended to use a Linux environment for optimal performance.

### Build

#### Build from source

1. Download the source code from this repository via Git or manually
2. Within the source code directory, open a terminal and run the following command to build the bot:

   ```bash
   go build -o djzth-utilities-bot main.go
   ```

3. This will compile the source code into an executable named `djzth-utilities-bot` that can be run on any Linux based system.

#### Build using Docker

1. Download the source code from this repository via Git or manually
2. Ensure you have Docker installed on your system. If not, you can follow the instructions on the [Docker website](https://docs.docker.com/get-docker/).
3. Navigate to the directory containing the Dockerfile in the source code directory.
4. Build the Docker image using the following command:

   ```bash
   docker build -t djzth-utilities-bot .
   ```

5. This will create a Docker image named `djzth-utilities-bot` that you can run as a container. You can verify that the image has been created successfully by running the following command:

   ```bash
   docker images
   ```

## Configuration

Before running the bot, you need to configure it with your Discord bot token and other settings. In this repo you will find a `config.example.json` file. You will need to copy this file and rename it to `config.json` and update it to contain all the necessary information for your bot to function properly. It needs to live in the same directory as the compiled binary or the source code if you are running it directly from there.

## Running the bot

This bot can be run ad-hoc via your terminal, but it was meant to run as a process in a docker container for ease of management and deployment. Below are the two methods to run the bot:

### Ad-hoc

To run the bot ad-hoc you can use either of the following commands in your terminal:

```bash
# Run the bot in the foreground
chmod +x ./djzth-utilities-bot
./djzth-utilities-bot

# Or run the bot from the uncompiled source code (great for development)
# Make sure you are in the directory where main.go is located
go run main.go
```

### Via Docker

1. Pull your image onto your host you will be running the bot on. If you built the image locally, you can skip this step. If you are pulling from a registry, use the following command:

   ```bash
   docker pull djzth-utilities-bot
   ```

2. Ensure you have a configured `config.json` file in your current directory with the correct configuration for your bot
3. Run the bot in a Docker container using the following command:

   ```bash
   docker run -d --restart-always --name djzth-utilities-bot -v $(pwd)/config.json:/config.json djzth-utilities-bot
   ```

   This command mounts the `config.json` file from your current directory to the `/config.json` path inside the container, allowing the bot to access its configuration.

## Hosting

For hosting the bot, you can use any Linux-based server or cloud service that supports Docker. In production, this bot is currently hosted on an OVH-Cloud Virtual Private Server (VPS) with the following specifications:

* **CPU**: 4 vCores
* **RAM**: 4 GB
* **Storage**: 80 GB SSD
* **Operating System**: Ubuntu 24.04 LTS
* **Network**: 500 Mbps

## Usage Documentation

### Slash Commands

* /addrole `<user>` `<role>`: Adds a specified role to a user. This command is only usable by users with roles under the `rolesRequiringApproval` in the config file. If the role to add is part of the `rolesRequiringApproval`, the command will send a request to the specified channel in the config file for approval before adding the role to the user. This is to ensure that only authorized users can assign certain roles.
* /removerole `<user>` `<role>`: Removes a specified role from a user. This command is also only usable by users with roles under the `rolesRequiringApproval` in the config file. If the role to remove is part of the `rolesRequiringApproval`, the command will send a request to the specified channel in the config file for approval before removing the role from the user. This ensures that only authorized users can remove certain roles.
* /listroles `<user>`: Lists all roles assigned to a specified user. This command is also only usable by users with roles under the `rolesRequiringApproval` in the config file
* /suggestion `<suggestion>` `<team>`: Submits a suggestion to the specified team. The suggestion will be sent to the appropriate channel based on the team specified, using the `leadershipChannelIds` list in the config. This command is available to all users in the server

### Menu commands

* Report Message: Users can report a message by using the elipses on the message, selectiong `Apps`, and then `Report Message`

### Audit Log

The bot will update a specific channel for audit logging purposes whenever:

* A member joins the server
* A member leaves the server
  * If the member has roles that are in the approvedRoles list, those roles are pinged in the accessControl channel
* Roles are added to a member
* A message is deleted
* A message is reported by the bot function

### Events

* When a member is giving the the `communityMembeRole`, the bot will welcome them in the `communityMemberGeneralChannelId`.
* When a guild invite ticket is opened, the bot will post a summary for inviters to easily copy/paste info into WoW
* When a new Death Jesters application is submitted, the bot will pin the embed, ping the `djsMemberRoleId`, and set a `djsAppLabel` tag on the forum post in `djsAppForumChannelId`
* When a user start streaming to Twitch, they are given the `Streaming Now` role and special section on the member list
* Removes embeds from specific channels under the `removeEmbedsFromChannels` list in the config file
* Notifies the user if they try to ping a restricted role in a message

### Caching

#### Message Caching

The bot will cache the most recent 1000 messages from the server. This is used by the audit log feature to track message deletions and report them in the appropriate channel. The cache will be cleared when the bot restarts

#### Member Caching

The bot will cache up to 3000 members in the server when it starts up. This is used to quickly access member information for commands and events. The cache will be cleared when the bot restarts
