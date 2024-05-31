# XamPress

<p align="center"><img src="https://socialify.git.ci/nonesubham/xampress/image?font=Source%20Code%20Pro&amp;language=1&amp;name=1&amp;owner=1&amp;pattern=Plus&amp;stargazers=1&amp;theme=Auto" alt="project-image"></p>

Xampress is a Command Line Interface (CLI) tool designed to simplify the management of local WordPress site installations on [XAMPP](https://www.apachefriends.org/) . By leveraging the power of [WP-CLI](https://github.com/wp-cli/wp-cli/) Xampress streamlines common tasks such as site setup configuration and maintenance making it easier for developers to work with local WordPress environments.


## Features

- Automated Site Installation: Quickly set up new WordPress sites on your local XAMPP server with a single command.
- Automated Site Deletion: Quickly delete WordPress sites from your local XAMPP server with a single command.
- Configuration Handling: Auto configure wp-config.php files and other essential settings.
- Cloning : Clone any existing site in one click. **(Working on)**
- Backup and Restore: Create backups of your site and restore them when needed. **(Soon)**
- Custom Domains: Instead of http://localhost/project-name use custom domain like http://project-name.local **(Soon)**


## Release Files


You can find suitable binary for your Operating System at **[Release](https://github.com/nonesubham/xampress/releases/)**.


## Build Your Own Binary
### Prerequisite
- Language : [GO](https://go.dev/doc/install)

<br>

**Clone this repo**
```
git clone github.com/nonesubham/xampress@latest
```
>Open terminal/console inside xampress

**Install all dependencies**
```
go mod tidy
```
**Check working or not**
```
go run main.go
```
If everything is fine you will get an error : 

> XamPress is not inside XAMP/LAMP directory, please follow installation guide. https://github.com/nonesubham/xampress/readMe.md

**Build your own binary**
```
go build
```
After executing this command `xampress` named binary will generate in project folder, if every step followed. then just follow **Installation** Step.
## Setup
Before starting the project, ensure you have installed the XAMPP server on your local machine by downloading it from the [XAMPP website](https://www.apachefriends.org/). Once XAMPP is installed, download the xampress binary. Next, navigate to the root directory of your XAMPP installation and create a new folder named `xampress`. Move the downloaded xampress binary into this newly created `xampress` folder. Now add this path to your `PATH` environment variable.

**For example :**

if you are a **Windows** user then you should have to add this path in your `PATH` environment variable.
>C:/xamp/xampress 
## Usage/Examples

**To create/delete new WordPress site/project :** 
``` 
xampress create/delete site-name
```

**To get list of all WordPress existing site/project :** 
``` 
xampress list
```

**To start/stop XAMPP server :** 
``` 
xampress server start/stop
```
**To add custom WordPress's user credentials or Database credentials :**
``` 
xampress config <flags>
```
***Available Flags :-***
>--sql-user : Set MySQL username

>--sql-pass : Set MySQL password

>--wp-user : Set WordPress username

>--wp-pass : Set WordPress password

>--wp-email : Set WordPress email

## TODO

- [ ] Existing Project/Site Cloning.
- [ ] Backup and Restore of existing projects.
- [ ] Custom domain mapping for each site/project.
- [ ] Token based login or no credential login.


## Support

For support, contact me on [Telegram](https://t.me/nonesubham) or raise issue on this repository.


