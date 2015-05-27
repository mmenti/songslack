(Episode 1 in my "Teach yourself Go" series)

A simple program that checks users' "I"m going"/"I'm interested in " Songkick events and posts these to a Slack channel.

It's probably awful, since it's the first thing I've ever written in Go, with no input from anyone who knows what they're doing...

This example uses Amazon's SimpleDB to store previously posted events, but you could of course replace this to support whatever datastore you wish to use.

I just add it to a cron job to regurlarly check for new event updates.

On Slack, it will look something like this, whenever someone clicks "I'm going" or "I'm interested" for a Songkick event:

<img src="https://www.evernote.com/shard/s1/sh/a59c9a75-fa60-4a6d-a1c7-f5c8a5ae813c/dd71a6061fe429cc/res/3f7ffde9-ea9b-4f41-a942-abe3348f1f99/skitch.png"/>

