# Bot Commands and Automation

This document describes the commands registered by the bot and the automated behavior currently wired into it. Features that depend on Discord channel, role, or team configuration will only work when those IDs and values are set in `config.yaml`.

## Commands

### `/ping`

Replies with `Pong!`.

### `/addrole`

Options:

| Option | Required | Description |
| --- | --- | --- |
| `target` | Yes | Member who will receive the role |
| `role` | Yes | Role to add |

The command is limited to members who have at least one role listed in `rolesRequiringApproval`. If the requested role is also listed there, the bot posts an approval request in `accessControlChannelId`; a member with `roleApproverId` can approve or deny it using the buttons. Other permitted requests are applied immediately. The requester receives a private response, and completed direct changes are logged in the access-control channel.

### `/removerole`

Options:

| Option | Required | Description |
| --- | --- | --- |
| `target` | Yes | Member from whom to remove the role |
| `role` | Yes | Role to remove |

Uses the same caller permission and approval rules as `/addrole`. Approval buttons are restricted to members with `roleApproverId`.

### `/listrole`

Option: `role` (required). Lists the members currently holding that role in a private response. The caller must have at least one role listed in `rolesRequiringApproval`.

### `/create-raid-team-info`

Creates a raid-team embed in `raidTeamsChannelId`. Required options are `team`, `game`, `schedule`, `current-prog`, `recruitment-contact`, `blurb`, and `currently-recruiting`; `app-link` is optional. Teams are offered from `raidTeams` in the configuration. Games are `World of Warcraft` and `Final Fantasy XIV`.

### `/update-raid-team-info`

Finds the existing bot-posted raid-team embed in `raidTeamsChannelId` and updates it. `team` is required; `game`, `schedule`, `current-prog`, `recruitment-contact`, `blurb`, `currently-recruiting`, and `app-link` are optional. Use `/create-raid-team-info` first if that team has no existing post.

### `/suggestion`

Options: `suggestion` (required text) and `team` (required destination). The destination choices are Rocket, Gravity, Phoenix, Ramrod, Death Jesters, Integrity, Eclipse, and Last Call, mapped to entries in `leadershipChannelIds`. The bot privately confirms submission and posts the suggestion embed in the configured leadership channel.

### `Report Message` (message context menu)

Use the message's context menu, then **Apps → Report Message**. The bot sends a link to the selected message to `moderationChannelId`, pings `moderatorRoleId`, and privately confirms the report to the reporter.

## Automated Behavior

### Self-service role menus

On startup, the bot posts selection menus in `roleSelectionChannelId` if a matching menu embed is not already present in the channel's latest 50 messages. Menu choices are drawn from `openRoles` by matching role names:

| Menu | Role-name matching |
| --- | --- |
| Raid and PvP Role Selection | `Raid`, `BGs`, `Arena`, or `RGB` |
| Mythic+ Key Selection | `Key` or `Delves` |
| Other Activities Selection | `Valor` or `Collectors` |
| Pronoun Selection | `They/Them`, `She/Her`, `He/Him`, `Other/Ask Me`, or `Any` |
| Game Role Selection | `Retail WoW` or `FFXIV` |

Choosing options adds the selected roles and removes unselected roles within that menu's category. Menu interactions receive a private summary. The raid/PvP, key, and activity menus allow multiple selections; pronoun and game menus allow one selection.

### Membership and role audit

- A member joining is logged in `auditLogChannelId`.
- A member leaving is logged there. If their cached roles include any role in `rolesRequiringApproval` or `approvedRoles`, those roles are also mentioned in `accessControlChannelId`.
- Restricted role additions and removals are logged in `auditLogChannelId`, including the target, role, and best available attribution. Roles listed in `openRoles` are excluded from these role-change logs.
- The bot caches up to 1,000 recent messages for deletion logging. A deletion is reported in `auditLogChannelId` with its channel, author, and text when cached; otherwise the log notes that the message was unavailable. Deletions by `ticketBotUserId` are skipped when the message is cached.
- Bot-issued direct role changes are tracked briefly so the audit entry can attribute the change to the command invoker.

### Community welcome and member count

When `communityMemberRole` is newly added to a member, the bot posts a randomized greeting in `communityMemberGeneralChannelId`. At startup, it counts cached members with that role and sets the bot's custom status to that count. The count is initialized during startup caching, not continuously refreshed by this handler.

### Death Jesters application posts

When a forum thread is created under `djsAppForumChannelId`, the bot pins the latest message, posts a mention of `djsMemberRoleId`, and applies the configured `djsAppLabel` tag. Each thread is processed once per bot process.

### ZtH invite tickets

For a thread created under `ticketChannelId`, the bot looks for a `zth-` thread and retries fetching its messages while waiting for one that contains at least two embeds and a mentioned user. It then posts a copy/paste embed containing the character, realm, a guild note, and an officer note. If the member has no server nickname, it posts a warning and pings `roleApproverId` when configured.

The generated **Ping Inviters** button pings `championRoleId` and can be used once per hour. **Sorry We Missed You** posts a message to the ticket user explaining how to request help again.

### Twitch streaming role

For presence updates in the configured guild, a member who has `streamTeamRoleId` receives `isLiveRoleId` while streaming on Twitch. The live role is removed when they stop streaming on Twitch or lose the stream-team role. Both role IDs must be configured.

### Embed suppression

For channels listed in `embedRemoveChannels`, the bot checks non-bot messages containing `http`; after a short delay, it suppresses any embeds Discord generated for that message.

### Health endpoint

The bot serves `GET /healthz` on port `8080`. It returns HTTP 200 while connected to Discord and during a 90-second reconnect grace period, then returns HTTP 503 if the gateway remains disconnected.

## Configuration Notes

Configure the feature-specific channel and role IDs in `config.yaml`. In particular, the streaming-role automation uses `streamTeamRoleId` and `isLiveRoleId`; these keys are not present in the current `config.example.yaml`. Team options for the raid-team commands come from `raidTeams`, and the suggestion destinations come from `leadershipChannelIds`.

The role-management permission model uses `rolesRequiringApproval` both to authorize command users and to decide which target roles require approval. `approvedRoles` is additionally used when deciding which roles to mention when a member leaves.
