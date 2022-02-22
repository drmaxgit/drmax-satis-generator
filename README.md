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


| Name       | Type                  | Description                                                                                     |
| ------------ | ----------------------- | ------------------------------------------------------------------------------------------------- |
| `--input`  | string (**required**) | Path to input file. Defaults to`input.json`. See `input.example.json` or below for instructions |
| `--output` | string (**required**) | Path to save the config file. Defaults to`satis.json`                                           |

## Input/Configuration file

See `input.example.json` for details. Effectively this can be [full blown satis.json config](https://getcomposer.org/doc/articles/handling-private-packages.md#setup) even with prefilled repositories. This config is then enhanced with contents of your private repositories when you run the script.

### Sources Configuration

On top of default satis configuration `sources` attribute is added which instruments the script how to obtain your repositories. The attribute is removed before saving the final file.
The sources attribute looks like this

```
{
  "name": "myorg/satis",
  # any other satis config
  "sources": [
    {
      "sourceType": "gitlab",
      "sourceIdent": "349181",
      "sourceAuth": "GITLAB_PAT"
    },
    #...other sources
  ]
}
```

The values of source can be of following values:


|                  | **sourceType** | **sourceIdent**                                                               | **sourceAuth**                                                                                                                                               |
| ------------------ | ---------------- | ------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Gitlab**       | `gitlab`       | _group id to process_<br>`349181` for https://gitlab.com/gitlab-examples      | _Personal Access Token_ - [Create one here](https://gitlab.com/-/profile/personal_access_tokens )<br>Required scopes are  `api, read_api, read_repository`   |
| **Github**       | `github`       | _organization slug_<br>`composer` for https://github.com/composer             | _Personal access token_ - [Create one here](https://github.com/settings/tokens )<br>Required scopes are  `repo, read_org`                                    |
| **Azure DevOps** | `azdo`         | _full organization + project path_<br>`https://dev.azure.com/myorg/myproject` | _Personal Access Token_<br>Link for creating access token is<br>`https://dev.azure.com/YOURORG/_usersSettings/tokens`<br> Required scopes are  `Code > read` |

## Known issues

- **Azure DevOps disabled repositories are not ommited** - Azure DevOps API adds the parameter in version 7.1 which is not currently available in the API adapter used.
