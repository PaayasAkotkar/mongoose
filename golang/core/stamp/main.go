// Package mstamp stamps down the handshake on the client machine
// on-stamp validation the download proceeds
package mstamp

// metal-model:
// Handshake Phase I:
// clause: gather complete info about the machine for security purpose
// reason: to help the malware teams easily detect the machine and file-a-complaint
// Handshake Phase II:
// clause: write down the custom comment on the machine and a security block
// reason: to help directly shutdown the machine's dirve entrily
// Handshake Phase II:
// clause: write down the requested file
// reason: confirmation that the current file is been written just need to start the downloading phase
