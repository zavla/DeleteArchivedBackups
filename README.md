## A DBA tool set for manipulating backup files. ##

The utility DeleteArchivedBackups is used to delete old databases backup files.
A DBA needs to keep files in a directory with different databases backup files rotated.

That is a DBA needs to to deleted all but several last files for each database.

Files must obey naming scheme.
Database name in a file name and file suffix define a group of database backups.
Every such group may have its most recent files and outdated files.

