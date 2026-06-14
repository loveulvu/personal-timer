import { Button, Form, Input, InputNumber, Modal, Space, message } from 'antd'
import { useEffect } from 'react'
import { api, DailyTask } from '../../api'
import { errorMessage } from '../../utils/labels'

type CompletionForm = {
  finishNote: string
  finishDescription: string
  actualMinutesOverride?: number | null
}

type TaskCompletionModalProps = {
  task: DailyTask | null
  mode: 'finish' | 'edit'
  onClose: () => void
  onSaved: () => Promise<void>
}

export function TaskCompletionModal({ task, mode, onClose, onSaved }: TaskCompletionModalProps) {
  const [form] = Form.useForm<CompletionForm>()

  useEffect(() => {
    if (!task) return
    form.setFieldsValue({
      finishNote: task.finish_note ?? '',
      finishDescription: task.finish_description ?? '',
      actualMinutesOverride: task.actual_seconds_override == null ? null : Math.round(task.actual_seconds_override / 60),
    })
  }, [task, form])

  async function submit(values: CompletionForm) {
    if (!task) return
    try {
      const input = {
        finish_note: values.finishNote.trim(),
        finish_description: values.finishDescription.trim(),
      }
      if (mode === 'finish') {
        await api.finishTask(task.id, input)
        message.success('任务已完成')
      } else {
        const durationUpdate = values.actualMinutesOverride == null
          ? { clear_actual_minutes_override: true }
          : { actual_minutes_override: values.actualMinutesOverride }
        await api.updateCompletedTask(task.id, {
          ...input,
          ...durationUpdate,
        })
        message.success('完成记录已更新')
      }
      onClose()
      await onSaved()
    } catch (err) {
      message.error(errorMessage(err))
    }
  }

  return (
    <Modal
      title={mode === 'finish' ? '完成任务' : '编辑记录'}
      open={task !== null}
      onCancel={onClose}
      footer={null}
      destroyOnHidden
    >
      <Form form={form} layout="vertical" onFinish={submit}>
        <Form.Item name="finishNote" label="完成备注" rules={[{ required: true, whitespace: true, message: '请输入完成备注' }]}>
          <Input placeholder="简要记录完成结果" />
        </Form.Item>
        <Form.Item name="finishDescription" label="完成描述" rules={[{ required: true, whitespace: true, message: '请输入完成描述' }]}>
          <Input.TextArea rows={4} placeholder="描述完成内容、问题或后续事项" />
        </Form.Item>
        {mode === 'edit' && (
          <Form.Item name="actualMinutesOverride" label="实际时长（分钟）" extra="清空后恢复使用计时会话聚合时长">
            <InputNumber min={0} precision={0} placeholder="使用计时会话时长" />
          </Form.Item>
        )}
        <Space>
          <Button type="primary" htmlType="submit">{mode === 'finish' ? '完成任务' : '保存'}</Button>
          <Button onClick={onClose}>取消</Button>
        </Space>
      </Form>
    </Modal>
  )
}
