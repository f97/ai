# One API Frontend Interface

This project is the frontend interface for One API, based on [Berry Free React Admin Template](https://github.com/codedthemes/berry-free-react-admin-template).

## Open Source Projects Used

The following open source projects are used as part of our project:

- [Berry Free React Admin Template](https://github.com/codedthemes/berry-free-react-admin-template)
- [minimal-ui-kit](minimal-ui-kit)

## 开发Description

当添加新的渠道hour，需要修改以下地方：

1. `web/berry/src/constants/ChannelConstants.js`

在该文件中的 `CHANNEL_OPTIONS` 添加新的渠道

```js
export const CHANNEL_OPTIONS = {
  //key 为渠道ID
  1: {
    key: 1, // 渠道ID
    text: "OpenAI", // 渠道名称
    value: 1, // 渠道ID
    color: "primary", // 渠道列表Show的颜色
  },
};
```

2. `web/berry/src/views/Channel/type/Config.js`

在该文件中的`typeConfig`添加新的渠道配置， 如果无需配置，可以不添加

```js
const typeConfig = {
  // key 为渠道ID
  3: {
    inputLabel: {
      // 输入框名称 配置
      // 对应的字段名称
      base_url: "AZURE_OPENAI_ENDPOINT",
      other: "Default API Version",
    },
    prompt: {
      // 输入框Hint 配置
      // 对应的字段名称
      base_url: "请填写AZURE_OPENAI_ENDPOINT",

      // 注意：通过判断 `other` 是否有值来判断是否需要Show `other` 输入框， Default是没有值的
      other: "Please enterDefaultAPIVersion，例如：2024-03-01-preview",
    },
    modelGroup: "openai", // Model组名称,这值是给 填入渠道支持Model 按钮使用的。 填入渠道支持Model 按钮会根据这值来获取Model组，如果填写Default是 openai
  },
};
```

## 许可证

本items目中使用的代码遵循 MIT 许可证。
