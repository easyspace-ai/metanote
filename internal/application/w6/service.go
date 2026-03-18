package w6

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"youmind-backend-v2/internal/domain/project"
	infraai "youmind-backend-v2/internal/infrastructure/ai"
)

// pagemakingInstruction is the fixed design spec sent to W6 pagemaker agent.
// It matches the Java INSTRUCTION in W6ApiServiceImpl (pagemaking_instruction).
const pagemakingInstructionPrefix = `您是一名具有高等教育经验的教学设计师和前端开发专家，对现代教学演示设计有深入理解，尤其擅长创建符合高校教学规范的交互式讲义。您的设计需兼顾知识体系的严谨性和教学呈现的直观性。
请根据提供的内容，设计一个符合中国高等院校教育教学风格和表达习惯的"中文" 可视化网页作品。

## 内容要求
- 采用学术性中文表述，符合学科专业术语规范
- 保持原文件的核心信息，但以更易读、可视化的方式呈现
- 涉及理论知识的内容，要概述性的讲解
- 必须包含 关键知识点的动画或交互仿真展示
- 必要时要添加如下内容，并做到非常详细的展开
  - 核心知识点标题（可采用疑问句式）
  - 关键理论/公式（突出显示，并详细解释关键参数定义与理论基础）
  - 教学案例/示意图（如果需要）
  - 思考题/延伸问题（底部固定区域，如果需要）
- 在页面底部添加作者信息区域，包含：
  - 版权信息: IECUBE Tutorial 2025
  - 页脚用较小字号和灰色字体声明，"本内容为人工智能生成，观点为转述原作者，不代表本公司立场，仅供参考和批判"

## 仿真动画要求
  - 使用JavaScript代码生成动画，确保动画流畅且易于理解
  - 动画必须与内容紧密相关，并有助于解释复杂概念
  - 动画应具有教育意义，能够帮助学习者更好地理解概念
  - 动画应遵循社会主义核心价值观，符合中国教育标准
  - 动画支持参数修改，有开始/结束等按钮调整参数实现仿真效果
  - 如需用到插件，请使用中国大陆可快速访问的cnd地址

## 公式要求
- 使用KaTeX进行公式渲染，确保公式正确显示
- 公式必须使用LaTeX语法书写，并确保公式在页面中正确渲染
- 公式必须使用Markdown语法进行标记，并确保公式在页面中正确渲染

## 设计风格
- 整体风格参考Linear App的简约现代设计
- 使用清晰的视觉层次结构，突出重要内容
- 配色方案应专业、和谐，适合长时间阅读

## 数据可视化
- 必要时用JavaScript代码生成图表来增强表达
- 数据需要忠实引用自原文，不要使用原文中不包含的数据
- 使用标准化图表：柱状图、折线图、比例图等，适当位置展示
- 图表配色应符合整体主题
- 每个图表包含清晰标题和数据来源
- 确保图表清晰可读，附有必要的解释文字
- 如需用到插件，请使用中国大陆可快速访问的cnd地址

## 交互体验
- 添加适当的微交互效果提升用户体验：
  - 按钮悬停时有轻微放大和颜色变化
  - 卡片元素悬停时有精致的阴影和边框效果
  - 页面滚动时有平滑过渡效果
  - 内容区块加载时有优雅的淡入动画

## 图标与视觉元素
- 使用专业图标库如Font Awesome或Material Icons
- 使用中国大陆可快速访问的cnd地址
- 根据内容主题选择合适的插图或图表展示数据
- 避免使用emoji作为主要图标

## 媒体资源
- 使用文档中的Markdown图片链接（如果有的话）
- 使用文档中的嵌入代码（如果有的话）

## 响应式设计
- 页面能够自适应在所有设备上（手机、平板、桌面）完美展示
 - 使用相对单位（如em、rem、vh、vw）而非固定像素值
 - 添加媒体查询，针对不同屏幕尺寸优化布局和字体大小
- 针对不同屏幕尺寸优化布局和字体大小
- 确保移动端有良好的触控体验
- 简化复杂组件：对于时间线、多列布局等复杂组件，确保它们能够自适应不同屏幕尺寸，必要时简化设计或提供替代布局。

## 技术规范
- 使用HTML5、TailwindCSS 3.0+和必要的JavaScript
- 专业图标库的展示通过CDN引入必要资源
- 图表展示时通过CDN引入chart.js， 并保证无错误
- 动画功能要保证动画可用(如果有动画)
- 实现完整的深色/浅色模式切换功能，默认跟随系统设置
- 代码结构清晰，包含适当注释，便于理解和维护
- 如需用到插件，请使用中国大陆可快速访问的cnd地址
- 注意，Tailwindcss 3.0+通过CDN引入的正确方式是:`
const pagemakingInstructionTailwind = "`<script src=\"https://cdn.tailwindcss.com\"></script>`"
const pagemakingInstructionSuffix = `
- 界面中引入的CDN链接必须保证中国大陆地区可访问性，如不可访问则使用国内镜像源

## 特别注意事项
- 测试指令：请在设计过程中模拟测试不同屏幕尺寸（特别是高度较小的屏幕），确保所有内容都能完整且优雅显示。

## 性能优化
- 确保页面加载速度快，避免不必要的大型资源
- 图片使用现代格式(WebP)并进行适当压缩
- 实现懒加载技术用于长页面内容


## 输出要求
- 提供完整可运行的单一HTML文件，包含所有必要的CSS和JavaScript
- 确保代码符合W3C标准，无错误警告
- 页面在不同浏览器中保持一致的外观和功能

请你像一个真正的网页设计专家一样思考，充分发挥你的专业技能和创造力，打造一个令人惊艳的HTML可视化网页作品！`

const pagemakingInstruction = pagemakingInstructionPrefix + pagemakingInstructionTailwind + pagemakingInstructionSuffix

// PageMakerService uses the W6 client + WebSocket bridge to generate rich HTML
// pages from an outline and attach them to a project as resources.
//
// This mirrors the old Java pipeline: genChat -> websocket events ->
// /interact/agent (pagemaker) -> /interact/artefact -> save HTML.
type PageMakerService struct {
	w6  *infraai.W6Client
	ws  *infraai.W6WS
	res project.ResourceRepository
}

func NewPageMakerService(
	w6 *infraai.W6Client,
	ws *infraai.W6WS,
	resRepo project.ResourceRepository,
) *PageMakerService {
	return &PageMakerService{
		w6:  w6,
		ws:  ws,
		res: resRepo,
	}
}

// ProgressFunc is called with progress steps during generation (for SSE streaming).
// Steps: created_chat, calling_pagemaker, waiting_artefact, got_artefact, saving, done.
type ProgressFunc func(step string)

// GeneratePageFromOutline triggers W6 pagemaker to create an HTML page for the
// given project and returns the persisted resource.
// If progress is provided, it is called at each step for real-time UI updates.
func (s *PageMakerService) GeneratePageFromOutline(
	ctx context.Context,
	projectID string,
	title string,
	knowledgePoints string,
	outline string,
	progress ...ProgressFunc,
) (*project.Resource, error) {
	var progressFn ProgressFunc
	if len(progress) > 0 {
		progressFn = progress[0]
	}
	if s.w6 == nil || s.ws == nil {
		return nil, fmt.Errorf("w6 pagemaker is not configured")
	}

	chatID, err := s.w6.StartChat(ctx)
	if err != nil {
		return nil, fmt.Errorf("start w6 chat: %w", err)
	}
	if progressFn != nil {
		progressFn("created_chat")
	}

	// Channel to receive the artefact id from the websocket listener.
	artefactCh := make(chan string, 1)
	errCh := make(chan error, 1)

	wsCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start websocket listener. We try to be liberal in what we accept: as long as
	// there is an "artefacts" array somewhere in the event with objects that have
	// an "id" field, we treat the first id as the generated artefact.
	go func() {
		err := s.ws.ConnectAndStream(wsCtx, chatID, func(ev infraai.W6Event) bool {
			// helper to scan any JSON blob for an artefacts[].id
			findArtefactID := func(raw json.RawMessage) string {
				if len(raw) == 0 {
					return ""
				}
				var generic map[string]any
				if err := json.Unmarshal(raw, &generic); err != nil {
					return ""
				}

				var walk func(v any) string
				walk = func(v any) string {
					switch x := v.(type) {
					case map[string]any:
						// direct artefacts array
						if arr, ok := x["artefacts"].([]any); ok && len(arr) > 0 {
							if first, ok := arr[0].(map[string]any); ok {
								if idv, ok := first["id"].(string); ok && idv != "" {
									return idv
								}
							}
						}
						for _, vv := range x {
							if id := walk(vv); id != "" {
								return id
							}
						}
					case []any:
						for _, vv := range x {
							if id := walk(vv); id != "" {
								return id
							}
						}
					}
					return ""
				}

				return walk(generic)
			}

			// Prefer artefacts in "current" snapshot if present, otherwise fall back to payload.
			if id := findArtefactID(ev.Current); id != "" {
				select {
				case artefactCh <- id:
				default:
				}
				return false
			}
			if id := findArtefactID(ev.Payload); id != "" {
				select {
				case artefactCh <- id:
				default:
				}
				return false
			}

			// activity-stop with no artefact means we stop listening.
			if ev.Type == "activity-stop" {
				return false
			}
			return true
		})
		if err != nil {
			errCh <- err
		}
	}()

	// Kick off the pagemaker agent.
	// pagemaking_instruction = fixed design spec (Java INSTRUCTION); optimize_instruction = user outline.
	payload := map[string]any{
		"title":                  title,
		"knowledge_points":       knowledgePoints,
		"pagemaking_instruction": pagemakingInstruction,
		"is_need_optimize":       "yes",
		"optimize_instruction":   outline,
	}

	if progressFn != nil {
		progressFn("calling_pagemaker")
	}
	if err := s.w6.CallAgent(ctx, chatID, "pagemaker", payload, ""); err != nil {
		return nil, fmt.Errorf("call w6 pagemaker: %w", err)
	}

	if progressFn != nil {
		progressFn("waiting_artefact")
	}
	// Wait for either artefact id or error / timeout.
	var artefactID string
	select {
	case artefactID = <-artefactCh:
		if progressFn != nil {
			progressFn("got_artefact")
		}
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("timeout waiting for w6 artefact")
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	cancel() // stop websocket listener

	if progressFn != nil {
		progressFn("saving")
	}
	art, err := s.w6.GetArtefact(ctx, artefactID)
	if err != nil {
		return nil, fmt.Errorf("get w6 artefact: %w", err)
	}

	// artefact.content is expected to be a JSON string. In practice W6 currently
	// returns the HTML as a base64-encoded string (e.g. "PCFET0NU..."), so we
	// decode when possible.
	var html string
	if len(art.Content) > 0 {
		if err := json.Unmarshal(art.Content, &html); err != nil {
			// Fallback: store raw JSON if content shape changed.
			html = string(art.Content)
		}

		// Try to base64-decode if it looks like encoded HTML.
		if html != "" && !strings.Contains(html, "<html") && !strings.Contains(html, "<!DOCTYPE") {
			if decoded, err := base64.StdEncoding.DecodeString(html); err == nil {
				decodedStr := string(decoded)
				if strings.Contains(decodedStr, "<html") || strings.Contains(decodedStr, "<!DOCTYPE") {
					html = decodedStr
				}
			}
		}
	}

	now := time.Now().UTC()
	res := &project.Resource{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		Type:      "html_page",
		Name:      title,
		Content:   &html,
		URL:       nil,
		Size:      nil,
		CreatedAt: now,
	}

	if err := s.res.Create(res); err != nil {
		return nil, fmt.Errorf("save resource: %w", err)
	}
	if progressFn != nil {
		progressFn("done")
	}
	return res, nil
}

