@internal/controller/password_put.go

- implement this
- use the same structure like in the other controllers, for example @internal/controller/password_post.go
- receives a payload of type data.PasswordBundle in the request
- for this purpose you will implement in the server a method that is called UpdatePasswordByUserId that takes the context, personId and the passworddata payload 


the server UpdatePasswordByUserId you do as follows:

- get salt and payload from database

general rule: ONCE SALT is set and stored in the database, not to be touched again!!!


```pseudocode
if salt is empty
   generate salt
   generate AES password 
   encrypt payload 
   update both salt and payload - the only situation when you update salt in the database, when it is empty
else 
   db payload must be filed out as well
   generate AES password from the salt and provided password
   try to decrypt _existent_ password payload with the AES
   if decrypt error, then return error. Once set, the password will not change, at least not in this version so you need to know the password that was used to successfully encrypt the first set payload

   if not decrypt error, then using the derived AES key, encrypt new payload and save it in the database. 
end if
```
