## A DBA tool set for manipulating backup files. ##
# If you are a DBA and supposed to create, manipulate and rotate backup files. #  

The utility DeleteArchivedBackups is used to delete old databases backup files.  
A DBA needs to keep files in a directory with different databases backup files rotated.  

That is a DBA needs to deleted all but several last files for each database.  

File names must obey a naming scheme - name_YYYY-MM-ddThh-mm-ss-sss-suffix.ext  

A database backup may have files with a 'FULL' suffix for full database backup and other suffices for differential backups.

For example you have several files:   
store1-2020-21-01T12-34-00-001-FULL.bak  
store1-2020-22-02T10-15-00-002-differential.bak  
store1-2020-23-01T09-02-00-001-FULL.bak  
store1-2020-23-04T01-01-00-001-differential.bak  
store1-2020-24-01T02-02-00-001-differential.bak  

What files from the example above are outdated files?  

