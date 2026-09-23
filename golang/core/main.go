// Package mongoose implements the tcp packets
// typically the pull resources
// mental-model:-
// ---------  request-made           ---------
// ---------       V                 ---------
// --------- first-handshake-machine ---------
// --------- plot the hole ---------> type A-> self cancel | type B -> packet download
// --------- push the request         ---------
// gathering:-
// 1. handshake machine info
// 2. handshake network info
package mongoose
