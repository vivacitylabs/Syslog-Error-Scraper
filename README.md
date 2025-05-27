# Syslog-Error-Scraper
for diagnosing on sensors with failing VPs - this script is run on a sensor which is/has been showing VP errors. By reading the full log history and checking for common errors, the user can quickly identify and fix the issue without having to read through pages of logs themselves.

# How to Use

to run the Syslog-Error-Scraper in its current state. firstly get onto the vivacitydevices server:

'''
ssh vivacitydevices
'''

we must then transfer the executable file from vivacitydevices, onto the device we want to check. Firstly open an rssh port on the target device in the usual Centricity way. Then the following command can be used 

'''
scp -P 4XXXX ~/robin/syslog2 ubunutu@localhost:~/
'''

then rssh onto the device the usual way and the Syslog-Scraper exevutable file should be in the home directory. To execute the file simply run the following:

'''
sudo ./syslog2
'''