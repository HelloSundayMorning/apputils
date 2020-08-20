# Push Notification




## Migrating data 

 - from Legacy to local service notification token table

```sql

select CONCAT('00000000-0000-0000-0000-',lpad(user_id::text, 12, '0')) as user_id, token, device_type as device_os, extract(epoch from created_at) * 1000000 created_at from device_registrations

```