*IRC Cloud bot*

this project uses go lang's websocket library to connect to [www.irccloud.com](http://www.irccloud.com).

this is just a base bot and does nothing except connecting to the server and print the response (if debug mode is enabled). One could build more interesting bots using this as the base.


**To build**

you first need to get the websocket module

    $ go get code.google.com/p/go.net/websocket

then clone this repo and cd to the code dir and use

    $ go build baseBot.go
    $ ./baseBot --email='yourEmail@here.info' --password='*****' --debug=true

***License***
Copyright (C) 2013 Akshay Shekher

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see http://www.gnu.org/licenses/.

Author(s): Akshay Shekher <voldyman666@gmail.com>