A small tool to download Logitech Circle videos. Uses undocumented APIs that'll probably go away at some point, and then I'll have to get an OAuth token. But today, it works.

Usage:
1. Download from the [Releases](https://github.com/pdbogen/circle/releases) page.
2. Run `circle -email=your-logitech-email -password=your-logitech-password`. A session token will be saved to `session.json` and a list of your Accessories (i.e., cameras) will be printed.
3. Run `circle -accessory=accessoryId -begin="Jan 01 2020 12:00:00" -duration=1h` to download any videos from noon until 1pm on Jan 1 2020. You can use `-end` instead of `-duration` if you want a specific time range. Videos will be saved in your working directory.

You can specify `-email` and `-password` on the second command, too; the `session.json` cache will be used unless it's expired; and if so, a new session token will automatically be obtained.

Questions/complaints/whatever, open an issue. Have a good!
