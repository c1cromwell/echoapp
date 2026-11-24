# Transparency report and warrant canary

Two public statements that tell users when something about Echo's position changes. Neither adds
secrecy. They exist so that silence means something.

**Status: templates only.** Nothing has been published. Both need a person who can speak for the
company to sign them, and counsel should read the canary wording before the first one goes out: a
canary is a legal instrument, and what a company may and may not say about an order it has
received depends on the order and the jurisdiction. I have not had that review done.

## What Echo can produce if asked

This is the honest core of both documents, and it is true only as far as
[`RELAY_METADATA_REVIEW.md`](../RELAY_METADATA_REVIEW.md) §3 and §4 say it is. Update it when that
review changes.

| Asked for | What exists |
|---|---|
| Message content | Nothing readable. Messages are end-to-end encrypted; Echo holds no keys |
| Who a person messaged, and when | For a delivered message: nothing. For a message still waiting for an offline recipient (up to 14 days): the waiting envelope, which today still names both parties (review §4.1) |
| A person's contacts | The contact list the person saved in the app, if any |
| Account details | The account's DID, username, public keys, device push tokens, creation time |
| Phone number | Only if the person turned on SMS recovery; see review finding 15 |
| IP address | Not stored by the application. The hosting provider and any proxy in front of the service may log it |
| Group membership | Yes: the server knows who is in a group |

## Transparency report (publish twice a year)

```
Echo transparency report — <period, e.g. 1 January to 30 June 2027>
Published <date> by <name, role>

Requests for user data received from governments or courts:   <number>
  of which Echo complied with, in whole or in part:           <number>
  of which Echo contested:                                    <number>
Accounts affected:                                            <number>
Requests to remove content or accounts:                       <number>
Requests we are not permitted to describe:                    <state what the law allows you to state>

What we were able to hand over, per request type: <one line each>
Changes since the last report to what Echo stores or can produce: <list, or "none">
```

## Warrant canary (publish monthly, on a fixed day)

```
As of <date>, <company name> has not:
  - received a government or court order that it is forbidden to disclose;
  - been required to build, weaken or change any part of Echo to enable access to user data;
  - handed over any encryption key, or been asked to;
  - had its servers or signing keys seized.

The next statement is due on <date>. If it does not appear, or appears without this section,
assume one of the above has changed.

Recent headline, to show this was not written in advance: "<headline>" — <publication>, <date>
```

How to publish it so that it means something:

1. Sign each statement with a key kept offline (the key ceremony is the natural place to create
   it) and publish the signature with the text.
2. Publish on a schedule and never backdate. A late canary is a tripped canary.
3. Keep every past statement available.
4. Decide now who signs when the usual signer is unavailable. A canary that lapses because
   someone was on holiday trains users to ignore it.
