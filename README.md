Introduction
------------

Net manager is a command line application for handling net manager paper work.

Compiling Net Manager
---------------------

Installing Tools
================

In order to compile net manager you will need golang compiler.
Download it from this link and install according to your operating system
instructions.

After installing the compiler you need to download the source code. You can do
it either by cloning the repo using git command line client. In this case you
will also need to install git command line client:
https://git-scm.com/downloads

Getting Source Code
===================

Once git command line client is installed you need to run:

```
$ git clone git@github.com:VictorDenisov/net_manager.git
```

Another option is to download zip file of the latest stable branch:

```
$ wget https://github.com/VictorDenisov/net_manager/archive/refs/heads/master.zip
```

Building Source Code
====================

In order to build source code simply run

```
$ go build
```

in the directory with net manager's source code.

```
$ go install
```

will move the compiled binary file into the bin directory of your golang
installation.

General Configuration
---------------------

Configuration files for net manager are stored in .net-manager directory in
your user's home directory.

In order to discover your home directory you can run net_manager and it
will complain about missing configuration directory in your home directory.

```
$ net_manager
Failed to read config from home dir: open /home/vdenisov/.net-manager/net-manager.conf: no such file or directory
Trying config file in the working directory
Failed to read config file:
open .net-manager.conf: no such file or directory
Proceeding without config file.
Parsed config: <nil>
Failed to read config from home dir: open /home/vdenisov/.net-manager/ContactListByName.csv: no such file or directory
Trying config file in the working directory
Failed to read call signs: Failed to open call signdb: ContactListByName.csv open ContactListByName.csv: no such file or directory%    
```

Configuration File
==================

The whole application is configured by net-manager.conf file in .net-manager
directory.

```
station:
    call: N6DVS
    signature: N6DVS
    mail:
        smtp-host: smtp.gmail.com
        port: 587
        password: <your password>
        email: denisovenator@gmail.com # replace with your email.
time-report:
    main-mail: nigelpgore@gmail.com
    cc-mail: K6SW@arrl.net
net-log-directory: /opt/Dropbox/radio/sjraces/net_manager/checkins
hospital-log-directory: /opt/Dropbox/radio/sjraces/net_manager/hospital
```

Following a Net
---------------

As you follow a net you can popu
