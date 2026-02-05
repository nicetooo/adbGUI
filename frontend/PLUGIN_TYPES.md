# 插件系统类型说明

## 类型体系概览

插件系统中存在两个不同但相关的 `Plugin` 类型，它们在不同的上下文中使用：

### 1. **插件数据模型** (`pluginStore.ts`)

表示完整的插件数据对象，包含从数据库读取的所有信息：

```typescript
interface Plugin {
  metadata: {           // 来自数据库（必需）
    id: string;         // ← 表单填写
    name: string;       // ← 表单填写
    version: string;    // ← 表单填写
    filters: {...};     // ← 表单填写
    config: {...};      // ← 表单填写
    // ...
  };
  sourceCode: string;   // 用户编写的代码（字符串）
  language: string;     // "typescript" | "javascript"
  compiledCode: string; // 编译后的 JS
}
```

**使用场景**：
- 前端 Store 状态管理
- 后端 API 返回的数据
- 数据库存储的结构

### 2. **插件脚本定义** (`plugin.d.ts`)

表示用户在代码编辑器中编写的插件对象结构：

```typescript
interface Plugin {
  metadata?: {...};    // 可选，会被后端忽略 ❌
  onEvent: (event, context) => PluginResult;  // ✅ 必需
  onInit?: (context) => void;                 // ✅ 可选
  onDestroy?: (context) => void;              // ✅ 可选
}
```

**使用场景**：
- Monaco 编辑器中的 TypeScript 类型检查
- 用户编写插件代码时的智能提示
- 代码补全和文档提示

## 数据流

```
┌──────────────────────────────────────────────────────────────┐
│ 1. 用户在表单填写 metadata                                     │
│    - id: "my-plugin"                                         │
│    - name: "My Plugin"                                       │
│    - filters: { sources: ["network"], ... }                 │
└──────────────────────────────────────────────────────────────┘
                        ↓ 保存到数据库
┌──────────────────────────────────────────────────────────────┐
│ 2. 用户在代码编辑器中编写逻辑                                   │
│    const plugin = {                                          │
│      onEvent: (event, context) => { ... }  // ← 只写逻辑     │
│    }                                                         │
└──────────────────────────────────────────────────────────────┘
                        ↓ 编译并保存
┌──────────────────────────────────────────────────────────────┐
│ 3. 后端组合完整的 Plugin 对象                                  │
│    {                                                         │
│      metadata: { ... },        // ← 从数据库读取             │
│      sourceCode: "...",        // ← 用户代码                 │
│      compiledCode: "..."       // ← 编译后的代码             │
│    }                                                         │
└──────────────────────────────────────────────────────────────┘
                        ↓ 返回给前端
┌──────────────────────────────────────────────────────────────┐
│ 4. 前端 Store 存储完整的 Plugin 对象                          │
│    - 列表展示：显示 metadata 信息                             │
│    - 编辑时：加载 sourceCode 到编辑器                         │
└──────────────────────────────────────────────────────────────┘
```

## 运行时行为

当插件执行时，后端的处理流程：

```go
// 1. 从数据库加载完整的 Plugin 对象
plugin := db.GetPlugin(id)  // 包含 metadata 和 sourceCode

// 2. 执行 compiledCode，提取函数引用
vm.RunString(plugin.CompiledCode)
pluginObj := vm.Get("plugin")

// 3. 只提取函数，不使用代码中的 metadata
plugin.OnEventFunc = pluginObj.Get("onEvent")    // ✅ 使用
plugin.OnInitFunc = pluginObj.Get("onInit")      // ✅ 使用
plugin.OnDestroy = pluginObj.Get("onDestroy")    // ✅ 使用

// ❌ 代码中的 metadata 被忽略，使用数据库中的 plugin.Metadata
```

## 为什么有两个 Plugin 类型？

1. **职责分离**：
   - 数据模型 Plugin：完整的插件数据（含 metadata、代码、编译结果）
   - 脚本定义 Plugin：用户代码中的插件对象（只含逻辑函数）

2. **类型安全**：
   - 数据模型确保后端返回的数据结构正确
   - 脚本定义确保用户编写的代码符合规范

3. **避免混淆**：
   - 通过文档和注释明确说明两者的区别和用途
   - 在不同的上下文中使用，不会同时出现

## 最佳实践

### ✅ 正确做法

```typescript
// 在代码编辑器中（plugin.d.ts 的 Plugin 类型）
const plugin: Plugin = {
  onEvent: (event, context) => {
    // 从 context.config 读取配置
    const threshold = context.config.slowThresholdMs || 2000;
    
    // 业务逻辑
    if (event.data.duration > threshold) {
      return { derivedEvents: [...] };
    }
    
    return { derivedEvents: [] };
  }
};
```

### ❌ 错误做法

```typescript
// 在代码编辑器中写 metadata（会被忽略）
const plugin: Plugin = {
  metadata: {                    // ❌ 无效！会被忽略
    id: "my-plugin",
    name: "My Plugin"
  },
  onEvent: (event, context) => {
    // ...
  }
};
```

### ✅ 正确配置方式

1. 在 **Basic Info** 页填写：id、name、version、author、description
2. 在 **Event Filters** 页配置：sources、types、urlPattern
3. 在 **Source Code** 页编写：onEvent、onInit、onDestroy

## 类型兼容性检查

| 字段 | 后端 Go | 前端 Store | plugin.d.ts |
|------|---------|-----------|-------------|
| metadata | PluginMetadata (必需) | metadata (必需) | metadata (可选/忽略) |
| sourceCode | string (必需) | sourceCode (必需) | N/A |
| language | string (必需) | language (必需) | N/A |
| compiledCode | string (必需) | compiledCode (必需) | N/A |
| onEvent | goja.Callable (运行时) | N/A | 函数定义 (必需) |
| onInit | goja.Callable (运行时) | N/A | 函数定义 (可选) |
| onDestroy | goja.Callable (运行时) | N/A | 函数定义 (可选) |

## 总结

- **前端 Store 的 `Plugin`** = 完整的数据模型（metadata 必需）
- **plugin.d.ts 的 `Plugin`** = 用户代码定义（metadata 可选/无效）
- **运行时使用的 metadata** = 来自数据库，不是来自代码
- **用户应该在表单中配置 metadata**，在代码中专注业务逻辑
