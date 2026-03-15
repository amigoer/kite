import type { Category } from '@/types/category'

/**
 * Mock 分类数据
 */
export const mockCategoryList: Category[] = [
  {
    id: '1a2b3c4d-0001-4000-a000-000000000001',
    name: '前端开发',
    slug: 'frontend',
    description: '涵盖 React、Vue、CSS、TypeScript 等前端技术栈的深度文章与实践指南。',
    postCount: 12,
    createdAt: '2025-12-01T10:00:00Z',
    updatedAt: '2026-03-10T14:20:00Z',
  },
  {
    id: '1a2b3c4d-0002-4000-a000-000000000002',
    name: '后端开发',
    slug: 'backend',
    description: '聚焦 Go、Gin、GORM、微服务架构等后端工程化实践。',
    postCount: 8,
    createdAt: '2025-12-01T10:00:00Z',
    updatedAt: '2026-03-08T09:15:00Z',
  },
  {
    id: '1a2b3c4d-0003-4000-a000-000000000003',
    name: 'DevOps',
    slug: 'devops',
    description: 'Docker、Kubernetes、CI/CD 流水线以及云原生基础设施的最佳实践。',
    postCount: 5,
    createdAt: '2026-01-15T08:00:00Z',
    updatedAt: '2026-03-03T11:30:00Z',
  },
  {
    id: '1a2b3c4d-0004-4000-a000-000000000004',
    name: 'AI',
    slug: 'ai',
    description: '大语言模型应用、Prompt Engineering、AI 工具链集成的探索与实践。',
    postCount: 3,
    createdAt: '2026-02-01T09:00:00Z',
    updatedAt: '2026-03-14T16:45:00Z',
  },
  {
    id: '1a2b3c4d-0005-4000-a000-000000000005',
    name: '数据库',
    slug: 'database',
    description: 'PostgreSQL、SQLite、Redis 等数据存储方案的选型、优化与运维经验。',
    postCount: 4,
    createdAt: '2026-01-20T13:00:00Z',
    updatedAt: '2026-02-28T10:00:00Z',
  },
  {
    id: '1a2b3c4d-0006-4000-a000-000000000006',
    name: '开源项目',
    slug: 'open-source',
    description: '开源社区参与经验、项目维护心得和值得关注的开源工具推荐。',
    postCount: 2,
    createdAt: '2026-02-10T15:00:00Z',
    updatedAt: '2026-03-01T08:30:00Z',
  },
]
