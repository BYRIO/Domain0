# Domain0

Domain0 是一个云资源管理系统，专为管理多个云服务商而设计。
它特别适合拥有众多资源和不同用户的组织。

- [功能](#功能)
  - [DNS 管理](#dns-管理)
  - [SSL 证书 管理](#ssl-证书-管理)
  - [IAM 身份与访问管理](#iam-身份与访问管理)
- [如何使用](#如何使用)
  - [我应该怎样获取 API key & secret](#我应该怎样获取-api-key--secret)
  - [身份权限说明](#身份权限说明)
  - [关于隐私](#关于隐私)

## 功能

### DNS 管理

Domain0 提供了一个 DNS 管理系统，使您可以在一个平台上管理多个域名的 DNS 记录，并可以授权其他用户进行访问。

- [x] 阿里云 DNS
- [x] 腾讯云 DNS
- [x] 华为云 DNS
- [x] Cloudflare DNS
- [ ] ...

欢迎贡献

> todo: dns 管理的 api key

### SSL 证书 管理

您可以在一个平台上为多个域名生成 SSL 证书，并将其分发给其他用户或上传至云端。

#### 签发

- [ ] Let's encrypt
- [ ] ZeroSSL
- [ ] 腾讯云 SSL
- [ ] 阿里云 SSL
- [ ] 华为云 SSL

#### 上传

- [ ] 阿里云 CDN
- [ ] 腾讯云 CDN
- [ ] 自建服务

### IAM 身份与访问管理

- [x] Feishu Oauth
- [ ] ...

## 安装

## 如何使用

### 我应该怎样获取 API key & secret

#### 阿里云

> api_id: **AccessKey ID**
> api_secret: **AccessKey Secret**

1. 访问 https://ram.console.aliyun.com/users/create 填写好信息后，勾选 OpenAPI 调用访问并创建用户。
2. 用户创建成功后提示的`AccessKey ID`和`AccessKey Secret`就是在本系统中所需要的 `api_id` 和 `api_secret`
3. 访问 https://ram.console.aliyun.com/permissions 即导航栏中的授权。
4. 点击新增授权按钮，为刚刚所创建的账号提供以下两个权限：`AliyunDNSFullAccess`、`AliyunHTTPDNSFullAccess`
   如果要更小范围的权限控制，可以在权限策略的
5. 回到本系统添加对应的域名。

#### 腾讯云

> TODO

#### 华为云

> TODO

#### Cloudflare

> api_id: **Zone ID**
> api_secret: \*\*\*\*

1. 访问你想添加的域名页面，即 https://dash.cloudflare.com/:id/:domain
2. 在 dashboard 的右侧获取你的 Zone ID
3. 点击下方的`Get your API token`，即 https://dash.cloudflare.com/profile/api-tokens
4. 设置好对应权限后，得到的 api token 就是本系统中的 api_secret
5. 回到本系统添加对应的域名。

### 身份权限说明

<table>
  <tr>
    <th>/</th>
    <th>域名创建</th>
    <th>域名管理</th>
    <th>权限管理</th>
    <th>补充说明</th>
  </tr>
  <tr style="text-align: center;">
    <td>Normal</td>
    <td>×</td>
    <td>×</td>
    <td>×</td>
    <td style="text-align: left;">
        新用户默认身份<br/>只允许访问被授予权限的域名<br/>进行域名记录修改的操作
    </td>
  </tr>
  <tr style="text-align: center;">
    <td>Contributor</td>
    <td>√</td>
    <td>只允许管理自己的域名</td>
    <td>允许其它用户访问自己可管理的域名</td>
    <td>-</td>
  </tr>
  <tr style="text-align: center;">
    <td style="text-align: left">Admin</td>
    <td>√</td>
    <td>允许管理所有域名</td>
    <td>允许其它用户访问自己可管理的域名<br/>可以提升或者降级用户的身分组（仅限Contributor及以下）</td>
    <td>-</td>
  </tr>
  <tr style="text-align: center;">
    <td>SysAdmin</td>
    <td colspan="2">与Admin一致</td>
    <td>与Admin一致<br/>可以提升或者降级用户的身分组（仅限Admin及以下）</td>
    <td>-</td>
  </tr>
</table>

### 关于隐私

在创建域名管理器时，您可以选择设置域名为**非公开**，这样 SysAdmin 也无法查看到你的域名信息。
