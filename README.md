# go_cli_zt

Status: Actively Supported.

## Purpose
This program is used to manage a self-hosted ZeroTier controller.

## Motivation
This is an update to the original program I wrote https://github.com/thedunston/bash_cli_zt. After I receoved from my doctoral dissertation, it was time to start working on bash_cli_zt again. However, I decided to switch to using Go.

The primary motivation for the switch to Go is to provide a CLI on multiple platforms and not having to manage multiple code bases. Initially, I was working on a ZT controller using PowerShell, but switched to Go for simpler maintenance on my end.

## One change from bash_cli_zt
One major change is that the "node.js" program is required to create Flow Rules. I decided not to try and recreate what the ZeroTier creator Adam Ierymenko has already developed. You'll need to download the static binary from: https://nodejs.org/download/nightly/v23.0.0-nightly20240512d78537b3df/

for your OS and then copy the 'node' program to the 'rules-compiler' folder once you clone this repo. On my tests with Windows and Linux, only the node.exe (windows) or the 'node' binary was required and not all the other files when using one of the static binaries.

**NOTE: On Windows, Windows Defender complained about the program because it does use system calls to clear the screen and execute the 'node.exe' program.**

## Web management

This version also has a very, very basic web interface that can be used to nanage the self-hosted controller, as well. For folks using Windows, docker can be a lot based on their system resources or folks who are using a Linux distro with minimal RAM so I wanted to provide another option for management.

The web interface features are similar to the CLI. I'll be adding more to that in the future.

Gemini helped me make it look like a terminal.

### Main Network Screen
![go_cli_zt_web](https://github.com/thedunston/go_cli_zt/assets/43048165/fe3c87ca-7977-44b3-87e9-d26e884d829c)

## Network Details Screen
![go_cli_zt_web2](https://github.com/thedunston/go_cli_zt/assets/43048165/f22f2320-6968-4302-a800-57732db13109)

## Setup Linux
```
git clone http://github.com/thedunston/go_cli_zt
cd go_cli_zt
chmod +x ztNetworks
sudo ./ztNetworks
```

`sudo` is required in order to view the ZeroTier Secrets file to query and POST to the controller.

## Setup Windows

```
git clone http://github.com/thedunston/go_cli_zt
cd go_cli_zt
.\ztNetworks.exe -cli
(or double-click on the ztNetworks.exe executable)
```
**REMINDER: Windows Defender may alert because system calls are made from the program.**

If you start the program without any CLI options or double-click, then it will open a terminal and start the web manager.



```
      ██████   ██████           ██████ ██      ██         ███████ ████████
     ██       ██    ██         ██      ██      ██            ███     ██
     ██   ███ ██    ██         ██      ██      ██           ███      ██
     ██    ██ ██    ██         ██      ██      ██          ███       ██
      ██████   ██████  ███████  ██████ ███████ ██ ███████ ███████    ██


                                 Duane Dunston
                              thedunston@gmail.com
Please send bug and feature requests here: https://github.com/thedunston/go_cli_zt

SUCCESS  Open your browser and connect to: http://localhost:4444

```

## Starting go_cli_zt
When you first run the program, it will prompt that it needs to create a SQLite database. That is where the `peers` and `networks` are stored for use with the CLI and web manager.

```

      ██████   ██████           ██████ ██      ██         ███████ ████████ 
     ██       ██    ██         ██      ██      ██            ███     ██    
     ██   ███ ██    ██         ██      ██      ██           ███      ██    
     ██    ██ ██    ██         ██      ██      ██          ███       ██    
      ██████   ██████  ███████  ██████ ███████ ██ ███████ ███████    ██    


                                 Duane Dunston
                              thedunston@gmail.com
Please send bug and feature requests here: https://github.com/thedunston/go_cli_zt

┌───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
|                                                                                                                   |
|                                                                                                                   |
|          goclzt needs to create and populate the SQLite database with the current ZT Networks and                 |
|          its peers.The database is located under: C:\Users\pinecone\AppData\Local\ZeroTier\wztPeerInfo.db         |
|                                                                                                                   |
|                                                                                                                   |
└───────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

Then you'll see the familiar interface with the same features as bash_cli_zt.

```
################################
#  ZeroTier Manager Controller
################################

1. Create a new ZT Network on this controller
2. Delete a ZT Network on this controller
3. Peer Management
4. Edit Flow Rules for Network
5. List all networks
6. Manage Routes
7. Update Network Description or IP Assignment
[E]xit
```

## AI Pair Programming

I used Gemini to help with parts of the program that required more brain power like dealing with CIDRs and start and end IPs. I learned a lot about JQuery with the web interface features and it generated the initial terminal theme. I am not that familiar with Javascript and CSS styling or designing is not a skill I have.
