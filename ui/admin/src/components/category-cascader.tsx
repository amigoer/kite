import { useState, useMemo, useCallback, useEffect } from 'react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import { ChevronDown, ChevronRight, Check, FolderOpen, X } from 'lucide-react'
import type { Category } from '@/types/category'

// ─── 类型定义 ────────────────────────────────────────────────────────

/** 级联选择器的节点数据结构 */
export interface CascaderNode {
  id: string
  name: string
  children?: CascaderNode[]
}

export interface CategoryCascaderProps {
  /** 树形数据源 */
  options: CascaderNode[]
  /** 当前选中的节点 ID（最终叶子节点 ID） */
  value?: string | null
  /** 选中回调，返回选中节点的 ID */
  onChange?: (id: string | null) => void
  /** 占位文字 */
  placeholder?: string
  /** 是否禁用 */
  disabled?: boolean
  /** 是否允许选择非叶子节点（父分类也可被选中） */
  allowSelectParent?: boolean
  /** 自定义触发器容器类名 */
  className?: string
}

// ─── 工具函数 ────────────────────────────────────────────────────────

/**
 * 根据节点 ID，通过深度优先搜索在树中查找完整路径
 * @param nodes 树形数据
 * @param targetId 目标节点 ID
 * @returns 从根到目标节点的路径数组，未找到返回 null
 */
function findPathById(nodes: CascaderNode[], targetId: string): CascaderNode[] | null {
  for (const node of nodes) {
    // 命中当前节点
    if (node.id === targetId) return [node]
    // 递归搜索子节点
    if (node.children?.length) {
      const childPath = findPathById(node.children, targetId)
      if (childPath) return [node, ...childPath]
    }
  }
  return null
}

/**
 * 将路径数组格式化为 「A / B / C」 的显示文本
 */
function formatPathLabel(path: CascaderNode[]): string {
  return path.map((n) => n.name).join(' / ')
}

/**
 * 将扁平的 Category 列表转换为 CascaderNode 树结构（支持无限层级）
 * @param flat 扁平分类数组（含 parentId 字段）
 * @returns 树形的 CascaderNode 数组
 */
export function buildCascaderTree(flat: Category[]): CascaderNode[] {
  // 先创建所有节点的映射
  const nodeMap = new Map<string, CascaderNode>()
  for (const cat of flat) {
    nodeMap.set(cat.id, { id: cat.id, name: cat.name })
  }

  const roots: CascaderNode[] = []

  // 遍历所有分类，将子节点挂载到父节点
  for (const cat of flat) {
    const node = nodeMap.get(cat.id)!
    if (cat.parentId) {
      const parent = nodeMap.get(cat.parentId)
      if (parent) {
        if (!parent.children) parent.children = []
        parent.children.push(node)
      } else {
        // parentId 指向不存在的节点，作为根节点处理
        roots.push(node)
      }
    } else {
      roots.push(node)
    }
  }

  return roots
}



// ─── 桌面端多列并排组件 ──────────────────────────────────────────────

interface ColumnPanelProps {
  options: CascaderNode[]
  value: string | null
  allowSelectParent: boolean
  onSelect: (id: string) => void
}

/**
 * 多列并排模式（macOS Finder 风格）
 * 选中父级后，右侧平滑展开子级列表
 */
function ColumnPanel({ options, value, allowSelectParent, onSelect }: ColumnPanelProps) {
  // 每一级选中的节点 ID
  const [activePath, setActivePath] = useState<string[]>([])

  // value 变化时，自动展开到对应路径
  useEffect(() => {
    if (value) {
      const path = findPathById(options, value)
      if (path) {
        setActivePath(path.map((n) => n.id))
        return
      }
    }
    setActivePath([])
  }, [value, options])

  /**
   * 构建待渲染的列集合
   * columns[0] = 顶层列表, columns[1] = 第一级选中节点的子列表, ...
   */
  const columns = useMemo(() => {
    const cols: { items: CascaderNode[]; activeId: string | null }[] = []
    let currentItems = options

    for (let depth = 0; depth <= activePath.length; depth++) {
      const activeId = activePath[depth] ?? null
      cols.push({ items: currentItems, activeId })

      // 找到当前级别选中的节点，取其 children 作为下一列
      if (activeId) {
        const activeNode = currentItems.find((n) => n.id === activeId)
        if (activeNode?.children?.length) {
          currentItems = activeNode.children
        } else {
          break
        }
      } else {
        break
      }
    }

    return cols
  }, [options, activePath])

  /** 处理某一级节点的 hover/click，更新 activePath */
  function handleNodeHover(depth: number, nodeId: string) {
    setActivePath((prev) => {
      const next = prev.slice(0, depth)
      next[depth] = nodeId
      return next
    })
  }

  return (
    <div
      className="flex overflow-hidden transition-all duration-200"
      style={{ width: `${columns.length * 200}px`, maxWidth: '600px' }}
    >
      {columns.map((col, colIdx) => (
        <div
          key={colIdx}
          className={cn(
            'w-[200px] shrink-0 animate-in slide-in-from-right-2 fade-in-50 duration-150',
            colIdx > 0 && 'border-l border-zinc-100 dark:border-zinc-800'
          )}
        >
          {/* 列标题 */}
          {colIdx === 0 && (
            <div className="px-3 py-1.5 border-b border-zinc-100 dark:border-zinc-800">
              <span className="text-[10px] text-zinc-400 font-medium uppercase tracking-wider">全部分类</span>
            </div>
          )}
          {colIdx > 0 && (
            <div className="px-3 py-1.5 border-b border-zinc-100 dark:border-zinc-800">
              <span className="text-[10px] text-zinc-400 font-medium uppercase tracking-wider">子分类</span>
            </div>
          )}

          <ScrollArea className="max-h-52">
            <div className="py-1">
              {col.items.length === 0 && (
                <div className="px-3 py-4 text-center text-xs text-zinc-400">暂无</div>
              )}
              {col.items.map((node) => {
                const isActive = col.activeId === node.id
                const isSelected = node.id === value
                const hasChildren = !!(node.children?.length)

                return (
                  <div
                    key={node.id}
                    className={cn(
                      'flex items-center gap-1.5 px-3 py-2 cursor-pointer transition-colors text-sm',
                      isSelected
                        ? 'bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900'
                        : isActive
                          ? 'bg-zinc-100 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100'
                          : 'text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800/60'
                    )}
                    onMouseEnter={() => handleNodeHover(colIdx, node.id)}
                    onClick={() => {
                      if (!hasChildren || allowSelectParent) {
                        onSelect(node.id)
                      } else {
                        // 有子级且不允许选父级，仅展开
                        handleNodeHover(colIdx, node.id)
                      }
                    }}
                  >
                    {isSelected && <Check className="w-3 h-3 shrink-0" />}
                    <span className="flex-1 truncate">{node.name}</span>
                    {hasChildren && <ChevronRight className="w-3 h-3 shrink-0 opacity-40" />}
                  </div>
                )
              })}
            </div>
          </ScrollArea>
        </div>
      ))}
    </div>
  )
}

// ─── 主组件 ──────────────────────────────────────────────────────────

/**
 * 分类级联选择器
 *
 * 基于 shadcn Popover：
 * - 桌面端：多列并排 (macOS Finder 风格)
 * - 触发器回显完整分类路径（例：后端 / Golang / 并发编程）
 *
 * @example
 * ```tsx
 * <CategoryCascader
 *   options={treeData}
 *   value={selectedId}
 *   onChange={setSelectedId}
 *   allowSelectParent
 * />
 * ```
 */
export function CategoryCascader({
  options,
  value = null,
  onChange,
  placeholder = '选择分类',
  disabled = false,
  allowSelectParent = true,
  className,
}: CategoryCascaderProps) {
  const [open, setOpen] = useState(false)

  /** 回显路径 label */
  const displayLabel = useMemo(() => {
    if (!value) return null
    const path = findPathById(options, value)
    if (!path) return null
    return formatPathLabel(path)
  }, [value, options])

  /** 选中处理 */
  const handleSelect = useCallback(
    (id: string) => {
      onChange?.(id)
      setOpen(false)
    },
    [onChange],
  )

  /** 清除选中 */
  const handleClear = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation()
      onChange?.(null)
    },
    [onChange],
  )

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn(
            'w-full justify-between font-normal shadow-none rounded-md border-zinc-200 dark:border-zinc-700 hover:bg-transparent',
            !value && 'text-muted-foreground',
            className,
          )}
        >
          <span className="flex items-center gap-1.5 truncate">
            {displayLabel ? (
              <>
                <FolderOpen className="w-3.5 h-3.5 text-zinc-400 shrink-0" />
                <span className="truncate">{displayLabel}</span>
              </>
            ) : (
              <span>{placeholder}</span>
            )}
          </span>
          <span className="flex items-center gap-0.5 shrink-0">
            {value && (
              <span
                className="p-0.5 rounded-sm hover:bg-zinc-100 dark:hover:bg-zinc-800 cursor-pointer"
                onClick={handleClear}
              >
                <X className="w-3 h-3 text-zinc-400" />
              </span>
            )}
            <ChevronDown className="w-3.5 h-3.5 text-zinc-400" />
          </span>
        </Button>
      </PopoverTrigger>
      <PopoverContent
        align="start"
        className="w-auto p-0 rounded-md"
      >
        {options.length === 0 ? (
          <div className="px-4 py-6 text-center text-xs text-zinc-400">暂无分类数据</div>
        ) : (
          <ColumnPanel
            options={options}
            value={value}
            allowSelectParent={allowSelectParent}
            onSelect={handleSelect}
          />
        )}
      </PopoverContent>
    </Popover>
  )
}
