(Episode 1 in my "Teach yourself Go" series)

A simple program that checks users' "I"m going"/"I'm interested in " Songkick events and posts these to a Slack channel.

It's probably awful, since it's the first thing I've ever written in Go, with no input from anyone who knows what they're doing...

This example uses Amazon's SimpleDB to store previously posted events, but you could of course replace this to support whatever datastore you wish to use.

I just add it to a cron job to regularly check for new event updates, and keep an eye on its output in case shit breaks.

It was originally writen for our own use in a music-oriented Slack team that happens to contain a lot of Songkick users, so we can easily see who is going to which shows (to keep it less noisy, we set up a separate #gigs channel for this). On Slack, it will look something like this, whenever someone clicks "I'm going" or "I'm interested" for a Songkick event:

<img src="https://www.evernote.com/shard/s1/sh/77d3b612-72b5-455e-b1a4-860cdf98d295/40f789cf50b6e4f6/res/540560a8-25a2-4b0d-8bbc-d059768d8b86/skitch.png"/>

