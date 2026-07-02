@internal/servr/server.go
 
- create in @internal/servr/server.go functions getPasswordData and updatePasswordData that return for a personID password salt and payload
- create function that generates salt for AES encryption, I understand it is 16 bytes if I am not wrong
- create function: starting from provided password, using salt which is that slice of 16 bytes, generate an AES key, there is some iterative process that is recommended to have about 600000 iterations
