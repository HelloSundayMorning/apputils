# Chat Channel

This package provides support for Chat Channel (CCS).

## Channel ID

A valid channel ID is in the format of `<channel role>-channel_<user ID>`.
E.g., `moderator-channel_00000000-0000-0000-0000-000000001234`

Refer to the source code [channel_id.go](./channel_id.go) for detailed info.


## Channel Type

A valid channel type is in the format of `<channel role>-channel`.

Currently `<channel role>` is either `carenav` or `moderator` representing the chat channel between a member and a care navigator and one between a member and a moderator.

Refer to the source code [channel_type.go](./channel_type.go) for detailed info.


## Channel Role

A "channel role" is a role that is teken by a user of the chat channel.

There are four predefined channel roles (`./channel_role.go`):
- `ChannelRoleCareNav` (`carenav`)
- `ChannelRoleModerator` (`moderator`)
- `ChannelRoleMember` (`member`)
- `ChannelRoleTriage` (`triage`)
All the other roles are represented as `ChannelRoleUnknown` (`unknown`)

Currently, only `carenav`, `moderator`, and `member` are used in the CCS-related services.

Refer to the source code [channel_role.go](./channel_role.go) for detailed info.


## User ID

A valid user ID is in the format of a UUID, but all its characters are digits (apart from the hyphen separators). Of all the digits, only the last few are non-zero values, the others are all zeros.
E.g., `00000000-0000-0000-0000-000000001234`

The non-zero value is the exact value which is used in the Legacy System as the old user ID.

Refer to the source code [uuid/uuid.go](../uuid/uuid.go) for detailed info.

