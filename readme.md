# Dolores

Simplifying secrets management on your cloud.

![architecture](./assets/architecture-small.png)

Encrypts configurations with different encryption algorithm and uses GCP or AWS or Vault as storage.

## Setup

Configure for different environments to manage secrets

```bash
dolores --env production init
```

Enter the GCS bucket name where you want to store the application configuration

## Encrypt

To encrypt a plain env file `backend.env` for production environments and upload it to GCS bucket, run the following

```bash
dolores --env production config encrypt -f backend.env --name backend-01
```
Once the file is encrypted successfully, you can remove the local plaintext file.

## Decrypt

To decrypt a remote encrypted file locally, run the following

```bash
dolores --environment production config decrypt --name backend-01 -key-file $HOME/.config/dolores/production.key
```

## Edit
You can edit the configuration file without need for decrypting, so it'll be updated remotely

```bash
dolores --environment production config edit --name backend-01 -key-file $HOME/.config/dolores/production.key
```
