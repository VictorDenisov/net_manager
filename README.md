Introduction
============

Net manager is a command line application for handling net manager paper work.

Compiling Net Manager
=====================

Installing Tools
----------------

In order to compile net manager you will need golang compiler.

Download it from [this](https://go.dev/dl/) link and install according to your
operating system instructions.

After installing the compiler you need to download the source code. You can do
it either by cloning the repo using git command line client. In this case you
will also need to install git command line client:
https://git-scm.com/downloads

Getting Source Code
-------------------

Once git command line client is installed you need to run:

```
$ git clone git@github.com:VictorDenisov/net_manager.git
```

Another option is to download zip file of the latest stable branch:

```
$ wget https://github.com/VictorDenisov/net_manager/archive/refs/heads/master.zip
```

Building Source Code
--------------------

In order to build source code simply run

```
$ go build
```

If you get errors on mac along the lines:
```
//go:linkname must refer to declared function or variable
```

Run this command

```
$ go get -u golang.org/x/sys
```

in the directory with net manager's source code.

```
$ go install
```

will move the compiled binary file into the bin directory of your golang
installation.

General Configuration
=====================

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
------------------

The whole application is configured by net-manager.conf file in .net-manager
directory.

```
# This line is a comment
station:
    call: N6DVS
    signature: N6DVS
    mail:
        smtp-host: smtp.gmail.com
        port: 587
        password: <your password>
        email: denisovenator@gmail.com
time-report:
    main-mail: chief_officer@gmail.com
    cc-mail: deputy_officer@arrl.net
net-log-directory: /opt/net_manager/checkins
hospital-log-directory: /opt/net_manager/hospital
mailing-list: main@ares-races.groups.io
```

In this configuration file you will need to replace N6DVS with your
station's call sign.

If you are using a different mailing server you will need to pass your
mailing server's smtp host. The port number will likely be the same.

Specify your smtp host password and your email. This email will be used
to cc you in all communications sent by the application.

Time reports are sent to SJ RACES chief radio officer - main-mail and secretary -
cc-mail.

You will also need ContactListByName.csv file from the membership database.
This database is used to find out names from call signs.

hospital_responsibility_schedule.txt file is a copy from: https://www.scc-ares-races.org/hospital/hospital-net-schedule.html
You need to keep this file up to date in order to make sure that
hospital net emails are populated properly.

city_responsibility_schedule.txt is downloaded from here: http://www.svecs.net/citynetcontroldates.html
This file is necessary to keep track of which city is net control of SVECS net.
You will need to keep this file up to date as well until automatic
retrieval of svecs schedule is implemented.

netcontrol_schedule.txt is the file that you update as people sign up
for net control positions.

Here is an example of net control schedule file. You only need to add records
to the file. Don't delete older records. They are necessary for generating
certificates at the end of the year.

```
07/05/2022	W6XRL4
07/12/2022	W6XRL3
07/19/2022	W6XRL1
07/26/2022	W6XRL8
08/02/2022	W6XRL2
08/09/2022	W6XRL3
08/16/2022	W6XRL9
```

Following a Net
===============

As you follow a net you can popu

```
$ net_manager -count -net-log 2022-09-13.txt
```

Sending Emails
==============

```
$ net_manager -send-emails
```

Generating Montly Timesheet
===========================

Most of the time you don't need to run this command manually, because it's
done by -send-emails command.
```
$ net_manager -time-sheet -month-prefix '2022-09'
```
