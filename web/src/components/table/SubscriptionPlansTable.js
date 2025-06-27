import React, { useEffect, useState } from 'react';
import { API, showError, showSuccess, renderNumber } from '../../helpers';

import {
  Package,
  DollarSign,
  Calendar,
  CheckCircle,
  XCircle,
  Settings,
  BarChart3,
  Users
} from 'lucide-react';
import {
  Button,
  Card,
  Divider,
  Dropdown,
  Empty,
  Form,
  Modal,
  Space,
  Table,
  Tag,
  Typography,
  Tooltip,
  Badge
} from '@douyinfe/semi-ui';
import {
  IllustrationNoResult,
  IllustrationNoResultDark
} from '@douyinfe/semi-illustrations';
import {
  IconPlus,
  IconSearch,
  IconEdit,
  IconDelete,
  IconStop,
  IconPlay,
  IconMore,
  IconRefresh
} from '@douyinfe/semi-icons';
import { ITEMS_PER_PAGE } from '../../constants';
import EditSubscriptionPlan from '../../pages/Subscription/EditSubscriptionPlan';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

const SubscriptionPlansTable = () => {
  const { t } = useTranslation();
  const [plans, setPlans] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searchStatus, setSearchStatus] = useState(-1);
  const [showEdit, setShowEdit] = useState(false);
  const [editingPlan, setEditingPlan] = useState(null);

  const columns = [
    {
      title: t('套餐名称'),
      dataIndex: 'name',
      render: (text, record) => (
        <Space>
          <Package size={16} />
          <Text strong>{text}</Text>
          {record.status === 1 && <Badge count="启用" type="success" />}
        </Space>
      ),
    },
    {
      title: t('价格'),
      dataIndex: 'price',
      render: (text) => (
        <Space>
          <DollarSign size={14} />
          <Text>¥{text}</Text>
        </Space>
      ),
    },
    {
      title: t('有效期'),
      dataIndex: 'duration',
      render: (text) => (
        <Space>
          <Calendar size={14} />
          <Text>{text}天</Text>
        </Space>
      ),
    },
    {
      title: t('模型配额'),
      dataIndex: 'model_quotas',
      render: (text) => {
        try {
          const quotas = JSON.parse(text || '{}');
          const quotaList = Object.entries(quotas);
          if (quotaList.length === 0) return '-';
          
          return (
            <Tooltip content={
              <div>
                {quotaList.map(([model, quota]) => (
                  <div key={model}>{model}: {quota}次</div>
                ))}
              </div>
            }>
              <Tag size="small">
                {quotaList.length}个模型
              </Tag>
            </Tooltip>
          );
        } catch (e) {
          return '-';
        }
      },
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      render: (text) => {
        return text === 1 ? (
          <Tag color='green' size='large' shape='circle' prefixIcon={<CheckCircle size={14} />}>
            {t('启用')}
          </Tag>
        ) : (
          <Tag color='red' size='large' shape='circle' prefixIcon={<XCircle size={14} />}>
            {t('禁用')}
          </Tag>
        );
      },
    },
    {
      title: t('创建时间'),
      dataIndex: 'created_time',
      render: (text) => {
        return new Date(text * 1000).toLocaleString();
      },
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      render: (text, record) => (
        <Dropdown
          trigger="click"
          position="bottomRight"
          menu={[
            {
              node: 'item',
              name: t('编辑'),
              icon: <IconEdit />,
              onClick: () => {
                setEditingPlan(record);
                setShowEdit(true);
              },
            },
            {
              node: 'item',
              name: record.status === 1 ? t('禁用') : t('启用'),
              icon: record.status === 1 ? <IconStop /> : <IconPlay />,
              onClick: () => {
                manageSubscriptionPlan(record.id, record.status === 1 ? 'disable' : 'enable');
              },
            },
            {
              node: 'item',
              name: t('删除'),
              icon: <IconDelete />,
              type: 'danger',
              onClick: () => {
                deleteSubscriptionPlan(record.id);
              },
            },
          ]}
        >
          <Button theme="borderless" icon={<IconMore />} />
        </Dropdown>
      ),
    },
  ];

  const setPlansFormat = (plans) => {
    setPlans(plans);
  };

  const loadPlans = async (startIdx = 0) => {
    setLoading(true);
    const res = await API.get(`/api/subscription/plans/page?page=${Math.floor(startIdx / ITEMS_PER_PAGE) + 1}&page_size=${ITEMS_PER_PAGE}&status=${searchStatus}`);
    const { success, message, data } = res.data;
    if (success) {
      setPlansFormat(data.plans);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const searchPlans = async () => {
    if (searchKeyword === '') {
      await loadPlans();
      return;
    }
    setLoading(true);
    const res = await API.get(`/api/subscription/plans/search?keyword=${searchKeyword}&status=${searchStatus}`);
    const { success, message, data } = res.data;
    if (success) {
      setPlansFormat(data);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const manageSubscriptionPlan = async (id, action) => {
    const res = await API.put(`/api/subscription/admin/plans/${id}`, {
      status: action === 'enable' ? 1 : 0
    });
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('操作成功'));
      await loadPlans();
    } else {
      showError(message);
    }
  };

  const deleteSubscriptionPlan = async (id) => {
    const res = await API.delete(`/api/subscription/admin/plans/${id}`);
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('删除成功'));
      await loadPlans();
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    loadPlans();
  }, []);

  const handlePageChange = (page) => {
    setActivePage(page);
    loadPlans((page - 1) * ITEMS_PER_PAGE);
  };

  return (
    <>
      <Card
        title={
          <Space>
            <Package size={20} />
            <Text size="large" strong>
              {t('订阅套餐管理')}
            </Text>
          </Space>
        }
        headerExtraContent={
          <Space>
            <Button
              theme="light"
              type="primary"
              icon={<IconRefresh />}
              onClick={() => loadPlans()}
            >
              {t('刷新')}
            </Button>
            <Button
              theme="solid"
              type="primary"
              icon={<IconPlus />}
              onClick={() => {
                setEditingPlan(null);
                setShowEdit(true);
              }}
            >
              {t('新建套餐')}
            </Button>
          </Space>
        }
      >
        <Space wrap style={{ marginBottom: 16 }}>
          <Form layout="horizontal" onSubmit={searchPlans}>
            <Form.Input
              field="keyword"
              placeholder={t('搜索套餐名称或描述')}
              value={searchKeyword}
              onChange={(value) => setSearchKeyword(value)}
              style={{ width: 200 }}
              suffix={<IconSearch />}
            />
            <Form.Select
              field="status"
              placeholder={t('状态')}
              value={searchStatus}
              onChange={(value) => setSearchStatus(value)}
              style={{ width: 120 }}
              optionList={[
                { label: t('全部'), value: -1 },
                { label: t('启用'), value: 1 },
                { label: t('禁用'), value: 0 },
              ]}
            />
            <Button type="primary" htmlType="submit" icon={<IconSearch />}>
              {t('搜索')}
            </Button>
          </Form>
        </Space>

        <Table
          columns={columns}
          dataSource={plans}
          pagination={{
            currentPage: activePage,
            pageSize: ITEMS_PER_PAGE,
            total: plans.length,
            showSizeChanger: true,
            onPageChange: handlePageChange,
          }}
          loading={loading}
          empty={
            <Empty
              image={<IllustrationNoResult />}
              darkModeImage={<IllustrationNoResultDark />}
              description={t('暂无数据')}
            />
          }
        />
      </Card>

      <Modal
        title={editingPlan ? t('编辑订阅套餐') : t('新建订阅套餐')}
        visible={showEdit}
        onCancel={() => setShowEdit(false)}
        footer={null}
        width={800}
      >
        <EditSubscriptionPlan
          plan={editingPlan}
          onCancel={() => setShowEdit(false)}
          onOk={() => {
            setShowEdit(false);
            loadPlans();
          }}
        />
      </Modal>
    </>
  );
};

export default SubscriptionPlansTable;
