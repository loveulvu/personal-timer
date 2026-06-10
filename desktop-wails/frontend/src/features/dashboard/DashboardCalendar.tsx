import { Card, Calendar } from 'antd'
import { Dayjs } from 'dayjs'

type DashboardCalendarProps = {
  date: Dayjs
  setDate: (date: Dayjs) => void
}

export function DashboardCalendar({ date, setDate }: DashboardCalendarProps) {
  return (
    <Card className="dashboard-card calendar-card" title="Calendar">
      <Calendar fullscreen={false} value={date} onSelect={setDate} />
    </Card>
  )
}
