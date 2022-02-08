# Dr.max Satis Generator
Prepares satis.json config based on contents of your Github/Gitlab/Azure DevOps organization repositories, fully customizable with input file

## Development
- `$ make dep` initiates go modules
- `$ make test` checks the code

### Important Dependencies
- https://github.com/google/go-github/v42/github used for Github communication
- https://github.com/microsoft/azure-devops-go-api used for AzDO communication
- https://github.com/xanzy/go-gitlab used for Gitlab communication
- https://gopkg.in/alecthomas/kingpin.v2 for cli input processing

## Run!
- `$ make` prepares win/linux/mac binaries into bin folder
- Use your preffered binary with following arguments

### Run Options

| Name              | Type                  | Description                                                                                       |
|-------------------|-----------------------|---------------------------------------------------------------------------------------------------|
| `--input`         | string (**required**) | Path to input file. Defaults to `input.json`. See `input.example.json` or below for instructions  |
| `--outpu`         | string (**required**) | Path to save the config file. Defaults to `satis.json`                                            |

## Input/Configuration file
See `input.example.json` for details. Effectively this can be [full blown satis.json config](https://getcomposer.org/doc/articles/handling-private-packages.md#setup) which is then enhanced with contents of your private repositories.

### Sources Configuration
On top of default satis configuration `sources` attribute is added which instruments the script how to obtain your repositories. The attribute is removed before saving the final file.
The sources attribute looks like this
```
{
  "name": "My Satis Repository",
  # any other satis config
  "sources": [
    {
      "sourceType": "gitlab",
      "sourceIdent": "349181",
      "sourceAuth": "GITLAB_PAT"
    },
    #other source
  ]
}
```

The values of source can be of following values:
- **sourceType** - `gitlab`, `github` or `azdo`
- **sourceIdent** based on value of `sourceType` either:
  - **gitlab** group id to process i.e. `349181` for https://gitlab.com/gitlab-examples
  - **github** organization slug i.e. `composer` for https://github.com/composer
  - **azdo** full organization + project path i.e. `https://dev.azure.com/myorg/myproject`
- **sourceAuth** based on value of `sourceType` either:
  - **gitlab** Personal Access Token. Create access token [here](https://gitlab.com/-/profile/personal_access_tokens). Required scopes are `api, read_api, read_repository`
  - **github** Personal access token. Create access token [here](https://github.com/settings/tokens). Required scopes are `repo, read_org`
  - **azdo** Personal Access Token. Link for creating access token is `https://dev.azure.com/myorg/_usersSettings/tokens`. Required scopes are `Code> read`

