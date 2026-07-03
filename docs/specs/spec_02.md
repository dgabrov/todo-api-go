@internal/servr/server.go

- implement again GetPasswordByUserId
- do not update anything in the database, this method is solely for accessing, reading
- it works as follows: 
1. load data
2. if salt and payload are not zero length, decode 64 bytes
3. if salt is zero length, must that payload is also zero length and then just return empty payload - with array not nill, as already shown there and do nothing more
4. if payload is not zero length, then salt should not be zero length as well, and then do as follows
   - derive the key using the salt
   - decrypt, parse the decrypted json to data.PasswordData and return
   - if decrypt is eroneous, then the error should show that the password is not good

   