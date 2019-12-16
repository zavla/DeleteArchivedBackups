## A DBA backup files tool set. ##

The utility DeleteArchivedBackups is used in databases backups operations.
Sometimes a DBA needs to keep a directory with different backup files rotated.
That is when a DBA needs to delete marked backup files. Or when one needs to deleted all but serveral last files.

Files must obey naming scheme.
Database name in file name and file suffix define a file group.
Every file group may have its most recent files and outdated files.
DeleteArchivedBackups deals with this task.
