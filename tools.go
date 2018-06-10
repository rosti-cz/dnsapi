package main

import (
	"golang.org/x/crypto/ssh"
	"fmt"
	"github.com/pkg/sftp"
	"os"
	"bytes"
	"io/ioutil"
)

func loadSSHKey(path string) []byte {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return content
}

func sshClient(server string) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod
	signer, err := ssh.ParsePrivateKey(loadSSHKey(config.SSHKey))
	if err != nil {
		return nil, err
	}
	authMethods = append(authMethods, ssh.PublicKeys(signer))

	var supportedCiphers = []string{
		"aes128-ctr", "aes192-ctr", "aes256-ctr",
		"aes128-gcm@openssh.com",
		"arcfour256", "arcfour128",
		"twofish256-cbc",
		"twofish-cbc",
		"twofish128-cbc",
		"blowfish-cbc",
		"3des-cbc",
		"arcfour",
		"cast128-cbc",
		"aes256-cbc",
		"aes128-cbc",
	}

	var sshConfig ssh.Config
	sshConfig.SetDefaults()
	sshConfig.Ciphers = supportedCiphers

	sshClientConfig := &ssh.ClientConfig{
		User:            config.SSHUser,
		Auth:            authMethods,
		Config:          sshConfig,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// open SSH connection
	// ssh app@alpha-node-4.rosti.cz -p 12360
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server, 22), sshClientConfig)

	return client, err
}

func SendFileViaSSH(server string, filename string, content string) error {
	client, err := sshClient(server)
	if err != nil {
		return err
	}
	defer client.Close()

	// open an SFTP session over an existing ssh connection.
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	// Open the file
	f, err := sftpClient.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(content))

	return err
}

func SendCommandViaSSH(server string, command string) (*bytes.Buffer, error) {
	client, err := sshClient(server)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// Get session
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var stdouterr bytes.Buffer

	session.Stderr = &stdouterr
	session.Stdout = &stdouterr

	err = session.Run(command)
	return &stdouterr, err
}