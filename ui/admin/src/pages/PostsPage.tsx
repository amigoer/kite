import { useState } from 'react'
import { useNavigate } from 'react-router'
import { Table, Button, Input, Tag, Select, Card, Typography, Pagination, Tooltip, Modal } from '@douyinfe/semi-ui'
import { IconSearch, IconPlus, IconEdit, IconDelete } from '@douyinfe/semi-icons'
import { usePosts, useDeletePost } from '@/hooks/use-posts'
import { useCategoryList } from '@/hooks/use-categories'
import type { Post, PostStatus } from '@/types/post'
import type { ColumnProps } from '@douyinfe/semi-ui/lib/es/table'
import type { TagColor } from '@douyinfe/semi-ui/lib/es/tag'

const { Title, Text } = Typography

/** 状态标签配置 */
const statusConfig: Record<PostStatus, { label: string; color: TagColor }> = {
  published: { label: '已发布', color: 'blue' },
  draft: { label: '草稿', color: 'grey' },
  archived: { label: '已归档', color: 'yellow' },
}

/** 状态筛选标签 */
const statusTabs = [
  { key: 'all' as const, label: '全部' },
  { key: 'published' as const, label: '已发布' },
  { key: 'draft' as const, label: '草稿' },
  { key: 'archived' as const, label: '已归档' },
]

/** 格式化日期 */
function formatDate(dateStr: string | null) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

/**
 * 文章管理页面 — Semi Table / Input / Select / Tag / Button
 */
export function PostsPage() {
  const navigate = useNavigate()
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<PostStatus | 'all'>('all')
  const [categoryId, setCategoryId] = useState('')
  const [page, setPage] = useState(1)
  const [selectedKeys, setSelectedKeys] = useState<string[]>([])
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; title: string } | 'batch' | null>(null)
  const pageSize = 10

  const { data, isLoading } = usePosts({ page, pageSize, keyword, status, categoryId: categoryId || undefined })
  const { data: categories } = useCategoryList()
  const deleteMutation = useDeletePost()

  const columns: ColumnProps[] = [
    {
      title: '标题', dataIndex: 'title',
      render: (_: unknown, record: Record<string, unknown>) => (
        <div
          style={{ cursor: 'pointer' }}
          onClick={() => navigate(`/posts/${record.id}/edit`)}
        >
          <Text strong style={{ fontSize: 14 }}>{record.title as string}</Text>
          <div style={{ marginTop: 2 }}>
            <Text type="tertiary" size="small" ellipsis={{ showTooltip: true }} style={{ maxWidth: 400 }}>
              {record.summary as string}
            </Text>
          </div>
        </div>
      ),
    },
    {
      title: '分类', dataIndex: 'category', width: 100,
      render: (val: { name: string } | null) =>
        val ? <Tag size="small" color="light-blue">{val.name}</Tag> : <Text type="tertiary" size="small">—</Text>,
    },
    {
      title: '状态', dataIndex: 'status', width: 90, align: 'center' as const,
      render: (val: PostStatus) => {
        const cfg = statusConfig[val]
        return <Tag color={cfg.color} size="small">{cfg.label}</Tag>
      },
    },
    {
      title: '更新日期', dataIndex: 'updatedAt', width: 120,
      sorter: (a: Record<string, string>, b: Record<string, string>) => new Date(a.updatedAt).getTime() - new Date(b.updatedAt).getTime(),
      render: (val: string) => <Text type="tertiary" size="small">{formatDate(val)}</Text>,
    },
    {
      title: '操作', width: 100, align: 'center' as const,
      render: (_: unknown, record: Record<string, unknown>) => (
        <div
          style={{ display: 'flex', gap: 4, justifyContent: 'center' }}
          onClick={(e) => e.stopPropagation()}
          onDoubleClick={(e) => e.stopPropagation()}
        >
          <Tooltip content="编辑" position="top">
            <Button icon={<IconEdit />} theme="borderless" size="small" onClick={() => navigate(`/posts/${record.id}/edit`)} />
          </Tooltip>
          <Tooltip content="删除" position="top">
            <Button
              icon={<IconDelete />}
              theme="borderless"
              type="danger"
              size="small"
              onClick={() => setDeleteTarget({ id: record.id as string, title: record.title as string })}
            />
          </Tooltip>
        </div>
      ),
    },
  ]

  const rowSelection = {
    selectedRowKeys: selectedKeys,
    onChange: (_: unknown, selectedRows?: Post[]) => {
      setSelectedKeys(selectedRows?.map((r) => r.id) || [])
    },
  }

  function handleDelete() {
    if (deleteTarget === 'batch') {
      // 逐个删除选中的文章
      Promise.all(selectedKeys.map((id) => deleteMutation.mutateAsync(id)))
        .then(() => setSelectedKeys([]))
    } else if (deleteTarget) {
      deleteMutation.mutate(deleteTarget.id)
    }
    setDeleteTarget(null)
  }

  return (
    <div>
      {/* 标题区 */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
        <div>
          <Title heading={4} style={{ marginBottom: 4 }}>文章管理</Title>
          <Text type="tertiary" size="small">管理你的所有博客文章</Text>
        </div>
        <Button icon={<IconPlus />} theme="solid" onClick={() => navigate('/posts/new')}>新建文章</Button>
      </div>

      {/* 状态标签快速筛选 */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <div style={{ display: 'flex', gap: 6 }}>
          {statusTabs.map((tab) => (
            <Tag
              key={tab.key}
              color={status === tab.key ? 'blue' : undefined}
              type={status === tab.key ? 'light' : 'ghost'}
              size="large"
              style={{ cursor: 'pointer', padding: '4px 14px' }}
              onClick={() => { setStatus(tab.key); setPage(1) }}
            >
              {tab.label}
            </Tag>
          ))}
        </div>

        {/* 批量操作 */}
        {selectedKeys.length > 0 && (
          <Button
            icon={<IconDelete />}
            type="danger"
            size="small"
            onClick={() => setDeleteTarget('batch')}
          >
            删除 {selectedKeys.length} 篇
          </Button>
        )}
      </div>

      {/* 搜索与筛选 */}
      <div style={{ display: 'flex', gap: 12, marginBottom: 16 }}>
        <Input
          prefix={<IconSearch />}
          placeholder="搜索文章标题、摘要…"
          value={keyword}
          onChange={(v) => { setKeyword(v); setPage(1) }}
          style={{ flex: 1 }}
          showClear
        />
        <Select
          value={categoryId || 'all'}
          onChange={(v) => { setCategoryId(v === 'all' ? '' : v as string); setPage(1) }}
          style={{ width: 160 }}
          prefix="分类"
        >
          <Select.Option value="all">全部</Select.Option>
          {categories?.map((c) => (
            <Select.Option key={c.id} value={c.id}>{c.name}</Select.Option>
          ))}
        </Select>
      </div>

      {/* 文章列表 */}
      <Card bodyStyle={{ padding: 0 }}>
        <Table
          columns={columns}
          dataSource={data?.items || []}
          loading={isLoading}
          pagination={false}
          rowKey="id"
          rowSelection={rowSelection}
          onRow={(record) => ({
            style: { cursor: 'pointer' },
            onDoubleClick: () => navigate(`/posts/${(record as unknown as Post).id}/edit`),
          })}
          empty={<Text type="tertiary" style={{ padding: 40, display: 'block', textAlign: 'center' }}>没有找到匹配的文章</Text>}
        />
      </Card>

      {/* 分页 */}
      {data && data.pagination.total > 0 && (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 16 }}>
          <Text type="tertiary" size="small">共 {data.pagination.total} 篇文章</Text>
          <Pagination total={data.pagination.total} pageSize={pageSize} currentPage={page} onPageChange={setPage} showSizeChanger />
        </div>
      )}

      {/* 删除确认弹窗 */}
      <Modal
        title={null}
        header={null}
        visible={deleteTarget !== null}
        onOk={handleDelete}
        onCancel={() => setDeleteTarget(null)}
        okType="danger"
        okText="确认删除"
        cancelText="取消"
        closeOnEsc
        centered
        width={420}
        bodyStyle={{ padding: '32px 32px 24px' }}
        footerStyle={{ padding: '12px 32px 24px', borderTop: 'none' }}
      >
        <div style={{ textAlign: 'center' }}>
          <div style={{
            width: 56, height: 56, borderRadius: '50%',
            background: 'rgba(239, 68, 68, 0.08)',
            display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
            marginBottom: 16,
          }}>
            <IconDelete style={{ fontSize: 24, color: '#ef4444' }} />
          </div>
          <Title heading={5} style={{ marginBottom: 8 }}>
            {deleteTarget === 'batch' ? '批量删除文章' : '删除文章'}
          </Title>
          <Text type="tertiary" style={{ lineHeight: 1.6 }}>
            {deleteTarget === 'batch'
              ? `确定要删除选中的 ${selectedKeys.length} 篇文章吗？此操作不可撤销。`
              : deleteTarget && `确定要删除「${deleteTarget.title}」吗？此操作不可撤销。`
            }
          </Text>
        </div>
      </Modal>
    </div>
  )
}
