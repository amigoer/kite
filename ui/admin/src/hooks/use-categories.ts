import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { mockCategoryList } from '@/mocks/categories'
import type { Category } from '@/types/category'

/** 模拟延迟 */
function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/** 内存中的分类列表副本，支持增删改 */
let categoriesStore = [...mockCategoryList]

/**
 * Mock 获取分类列表
 */
async function fetchCategoryList(keyword?: string): Promise<Category[]> {
  await delay(200)
  let result = [...categoriesStore]
  if (keyword) {
    const kw = keyword.toLowerCase()
    result = result.filter(
      (c) => c.name.toLowerCase().includes(kw) || c.slug.toLowerCase().includes(kw)
    )
  }
  // 按文章数倒序
  result.sort((a, b) => b.postCount - a.postCount)
  return result
}

/**
 * Mock 创建分类
 */
async function createCategory(data: Pick<Category, 'name' | 'slug' | 'description'>): Promise<Category> {
  await delay(300)
  const newCat: Category = {
    id: crypto.randomUUID(),
    name: data.name,
    slug: data.slug,
    description: data.description,
    postCount: 0,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }
  categoriesStore = [newCat, ...categoriesStore]
  return newCat
}

/**
 * Mock 删除分类
 */
async function deleteCategory(id: string): Promise<void> {
  await delay(200)
  categoriesStore = categoriesStore.filter((c) => c.id !== id)
}

/**
 * 获取分类列表 Hook
 */
export function useCategoryList(keyword?: string) {
  return useQuery({
    queryKey: ['categoryList', keyword],
    queryFn: () => fetchCategoryList(keyword),
  })
}

/**
 * 创建分类 Hook
 */
export function useCreateCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categoryList'] })
    },
  })
}

/**
 * 删除分类 Hook
 */
export function useDeleteCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categoryList'] })
    },
  })
}
