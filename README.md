# Follower

Follower is a small command line app that helps you keep track of what you're working on and what you've already done.

Built to work with Jira and Git, Follower keeps a local record of your progress and the time you've spent on a task. It does not track key presses or monitor your activity. It is closer to a journaling tool for development work, with optional sync points into the tools you already use.

## Why?

It is frustrating, and often time consuming, to step away from where the work is happening just to update a ticket, log progress, or record time spent on a task.

Follower keeps that workflow in the terminal and brings those actions together in one place.

## How?

Follower tracks work against a Jira ticket, whether you're inside a Git repository or not.

`follower follow [TICKET-ID]` sets the current task as your active context.

When you're ready to capture progress, run `follower checkpoint`. This opens an editor where you can write notes about what you've done.

If that checkpoint should also be prepared for Git or Jira, you can pass `--git`, `--jira`, or both. Follower will create a local commit for Git checkpoints and store Jira-related progress locally so it is ready to sync later.

When the work is ready to publish, the intended flow is to run `follower sync`, which will push Git changes and send the related Jira updates.
