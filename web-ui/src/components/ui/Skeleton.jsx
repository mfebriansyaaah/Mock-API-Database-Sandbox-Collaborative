import { cn } from '@/utils/cn.js'

export default function Skeleton({ className, ...rest }) {
  return <div className={cn('skeleton', className)} {...rest} />
}
