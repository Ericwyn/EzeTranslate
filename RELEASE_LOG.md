# EzeTranslate Release Log

## V1.1
- 发布：2021-11-17
- 说明：
  - 1. 更新迷你模式
    - 只显示翻译结果窗口，通过快捷键直接获取划词并显示翻译结果
  - 2. 有道翻译自动识别输入和输出语言
  
## V1.0
- 发布：2021-11-14
- 说明：
  - 1. 首个测试版本，实现基本功能功能
  - 2. 支持多个翻译接口，可便捷切换
    - 支持 Baidu 翻译 API
      - 需要在参数设置当中配置 Baidu API 参数
    - 支持 Google 翻译 API
      - 国内的话, 需要在参数设置当中配置 translate.google.com 的代理
      - 可以使用 cloudflare worker 做 js proxy
    - 支持 Youdao 翻译 API
      - 需要在参数设置当中配置 Youdao API 参数
  - 3. 针对代码注释进行格式化
    - 去除 `//` `/*` `#` 之类的符号
    - 去除回车, 多余空格等
  - 4. 划词翻译 (仅支持 Linux, 需安装 xclip )
