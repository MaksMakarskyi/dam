# About
Comming soon...

TODO:
- [x] Implement the first algorithm with the sophisticated frame, interface, and file structure
- [ ] Implement more algorithms (might change):
    - [ ] Leaky Bucket
    - [ ] Token Bucket
    - [x] Fixed Window Counter
    - [ ] Sliding Window Log
    - [ ] Sliding Window Counter
- [x] Implement common key functions to limit by (might change):
    - [x] IP address
    - [x] User ID (see JWTClaim key function)
    - [x] API key
    - [x] Session (see JWTClaim key function)
    - [x] Common key (global limiter, all requests have the same key)
- [ ] Improve the API of middleware and limiter packages to fit nicely with:
    - [ ] net/http
    - [ ] Chi 
    - [ ] Gin
    - [ ] Fiber
- [ ] Write sophisticated documentation for the main components
- [ ] Write README with tutorials