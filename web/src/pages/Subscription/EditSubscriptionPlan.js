import React, { useEffect, useState } from 'react';
import { API, showError, showSuccess } from '../../helpers';
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Select,
  Space,
  Typography,
  Divider,
  Tag,
  Modal
} from '@douyinfe/semi-ui';
import { Package, DollarSign, Calendar, Settings, Plus, Trash2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';

const { Text, Title } = Typography;

const EditSubscriptionPlan = ({ plan, onCancel, onOk }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [formApi, setFormApi] = useState();
  const [modelQuotas, setModelQuotas] = useState([]);

  const isEdit = !!plan;

  useEffect(() => {
    if (plan) {
      // 编辑模式，填充表单数据
      const quotas = JSON.parse(plan.model_quotas || '{}');
      const quotaList = Object.entries(quotas).map(([model, quota]) => ({
        model,
        quota: parseInt(quota)
      }));
      setModelQuotas(quotaList);
      
      if (formApi) {
        formApi.setValues({
          name: plan.name,
          description: plan.description,
          price: plan.price,
          duration: plan.duration,
          status: plan.status
        });
      }
    } else {
      // 新建模式，重置表单
      setModelQuotas([]);
      if (formApi) {
        formApi.reset();
      }
    }
  }, [plan, formApi]);

  const handleSubmit = async (values) => {
    if (modelQuotas.length === 0) {
      showError(t('请至少添加一个模型配额'));
      return;
    }

    // 验证模型配额
    for (const quota of modelQuotas) {
      if (!quota.model || quota.quota <= 0) {
        showError(t('请填写有效的模型名称和配额'));
        return;
      }
    }

    setLoading(true);
    
    const quotasObj = {};
    modelQuotas.forEach(({ model, quota }) => {
      quotasObj[model] = quota;
    });

    const data = {
      ...values,
      model_quotas: quotasObj
    };

    try {
      let res;
      if (isEdit) {
        res = await API.put(`/api/subscription/admin/plans/${plan.id}`, data);
      } else {
        res = await API.post('/api/subscription/admin/plans', data);
      }

      const { success, message } = res.data;
      if (success) {
        showSuccess(isEdit ? t('更新成功') : t('创建成功'));
        onOk();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const addModelQuota = () => {
    setModelQuotas([...modelQuotas, { model: '', quota: 100 }]);
  };

  const removeModelQuota = (index) => {
    const newQuotas = modelQuotas.filter((_, i) => i !== index);
    setModelQuotas(newQuotas);
  };

  const updateModelQuota = (index, field, value) => {
    const newQuotas = [...modelQuotas];
    newQuotas[index][field] = value;
    setModelQuotas(newQuotas);
  };

  const commonModels = [
    'gpt-4',
    'gpt-4-turbo',
    'gpt-4o',
    'gpt-4o-mini',
    'gpt-3.5-turbo',
    'claude-3-opus',
    'claude-3-sonnet',
    'claude-3-haiku',
    'claude-3.5-sonnet',
    'gemini-pro',
    'gemini-1.5-pro',
    'gemini-1.5-flash'
  ];

  return (
    <Form
      getFormApi={(api) => setFormApi(api)}
      onSubmit={handleSubmit}
      labelPosition="left"
      labelWidth={100}
    >
      <Card>
        <Space vertical style={{ width: '100%' }} spacing={16}>
          <div>
            <Title heading={6}>
              <Package size={16} style={{ marginRight: 8 }} />
              {t('基本信息')}
            </Title>
            <Divider margin={12} />
            
            <Form.Input
              field="name"
              label={t('套餐名称')}
              placeholder={t('请输入套餐名称')}
              rules={[{ required: true, message: t('套餐名称不能为空') }]}
              style={{ width: '100%' }}
            />
            
            <Form.TextArea
              field="description"
              label={t('套餐描述')}
              placeholder={t('请输入套餐描述')}
              rows={3}
              style={{ width: '100%' }}
            />
            
            <Form.InputNumber
              field="price"
              label={t('价格')}
              placeholder={t('请输入价格')}
              rules={[{ required: true, message: t('价格不能为空') }]}
              min={0}
              precision={2}
              prefix={<DollarSign size={14} />}
              suffix="元"
              style={{ width: '100%' }}
            />
            
            <Form.InputNumber
              field="duration"
              label={t('有效期')}
              placeholder={t('请输入有效期')}
              rules={[{ required: true, message: t('有效期不能为空') }]}
              min={1}
              prefix={<Calendar size={14} />}
              suffix="天"
              style={{ width: '100%' }}
            />
            
            <Form.Select
              field="status"
              label={t('状态')}
              placeholder={t('请选择状态')}
              rules={[{ required: true, message: t('请选择状态') }]}
              style={{ width: '100%' }}
              optionList={[
                { label: t('启用'), value: 1 },
                { label: t('禁用'), value: 0 }
              ]}
            />
          </div>

          <div>
            <Title heading={6}>
              <Settings size={16} style={{ marginRight: 8 }} />
              {t('模型配额')}
            </Title>
            <Divider margin={12} />
            
            <Space vertical style={{ width: '100%' }} spacing={12}>
              {modelQuotas.map((quota, index) => (
                <Card key={index} size="small" style={{ background: '#f8f9fa' }}>
                  <Space style={{ width: '100%' }}>
                    <Select
                      placeholder={t('选择模型')}
                      value={quota.model}
                      onChange={(value) => updateModelQuota(index, 'model', value)}
                      style={{ width: 200 }}
                      filter
                      optionList={commonModels.map(model => ({ label: model, value: model }))}
                    />
                    <InputNumber
                      placeholder={t('配额数量')}
                      value={quota.quota}
                      onChange={(value) => updateModelQuota(index, 'quota', value)}
                      min={1}
                      suffix="次"
                      style={{ width: 150 }}
                    />
                    <Button
                      type="danger"
                      theme="borderless"
                      icon={<Trash2 size={14} />}
                      onClick={() => removeModelQuota(index)}
                    />
                  </Space>
                </Card>
              ))}
              
              <Button
                type="dashed"
                icon={<Plus size={14} />}
                onClick={addModelQuota}
                style={{ width: '100%' }}
              >
                {t('添加模型配额')}
              </Button>
            </Space>
          </div>
        </Space>
      </Card>

      <div style={{ marginTop: 16, textAlign: 'right' }}>
        <Space>
          <Button onClick={onCancel}>
            {t('取消')}
          </Button>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
          >
            {isEdit ? t('更新') : t('创建')}
          </Button>
        </Space>
      </div>
    </Form>
  );
};

export default EditSubscriptionPlan;
