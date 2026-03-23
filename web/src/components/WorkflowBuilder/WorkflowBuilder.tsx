import React, { useCallback, useState } from 'react';
import ReactFlow, {
  Node,
  Edge,
  addEdge,
  Connection,
  useNodesState,
  useEdgesState,
  Controls,
  MiniMap,
  Background,
  NodeTypes,
} from 'reactflow';
import 'reactflow/dist/style.css';

import { NodePalette } from './NodePalette';
import { PropertiesPanel } from './PropertiesPanel';
import { BaseNode } from './nodes/BaseNode';
import { nodeTypes } from './nodeTypes';
import { useWorkflowStore } from '@/stores/workflowStore';
import { Button } from '@/components/ui/button';
import { Save, Play, RotateCcw } from 'lucide-react';

interface WorkflowBuilderProps {
  runbookId?: string;
  initialNodes?: Node[];
  initialEdges?: Edge[];
  onSave?: (workflow: any) => void;
  onExecute?: (workflow: any) => void;
}

export const WorkflowBuilder: React.FC<WorkflowBuilderProps> = ({
  runbookId,
  initialNodes = [],
  initialEdges = [],
  onSave,
  onExecute,
}) => {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [selectedEdge, setSelectedEdge] = useState<Edge | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [isExecuting, setIsExecuting] = useState(false);
  
  const { updateWorkflow, validateWorkflow } = useWorkflowStore();

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  );

  const onNodeClick = useCallback((event: React.MouseEvent, node: Node) => {
    setSelectedNode(node);
    setSelectedEdge(null);
  }, []);

  const onEdgeClick = useCallback((event: React.MouseEvent, edge: Edge) => {
    setSelectedEdge(edge);
    setSelectedNode(null);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
    setSelectedEdge(null);
  }, []);

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();

      const type = event.dataTransfer.getData('application/reactflow');
      if (typeof type === 'undefined' || !type) {
        return;
      }

      const reactFlowBounds = event.currentTarget.getBoundingClientRect();
      const position = {
        x: event.clientX - reactFlowBounds.left,
        y: event.clientY - reactFlowBounds.top,
      };

      const newNode: Node = {
        id: `${type}-${Date.now()}`,
        type,
        position,
        data: {
          label: `${type} Node`,
          config: getDefaultNodeConfig(type),
        },
      };

      setNodes((nds) => nds.concat(newNode));
    },
    [setNodes]
  );

  const handleSave = useCallback(async () => {
    setIsSaving(true);
    try {
      const isValid = validateWorkflow({ nodes, edges });
      if (!isValid) {
        throw new Error('Workflow validation failed');
      }

      const workflow = {
        id: runbookId,
        definition: {
          nodes,
          edges,
        },
      };

      if (onSave) {
        await onSave(workflow);
      }
    } catch (error) {
      console.error('Failed to save workflow:', error);
    } finally {
      setIsSaving(false);
    }
  }, [nodes, edges, runbookId, onSave, validateWorkflow]);

  const handleExecute = useCallback(async () => {
    setIsExecuting(true);
    try {
      const isValid = validateWorkflow({ nodes, edges });
      if (!isValid) {
        throw new Error('Workflow validation failed');
      }

      const workflow = {
        id: runbookId,
        definition: {
          nodes,
          edges,
        },
      };

      if (onExecute) {
        await onExecute(workflow);
      }
    } catch (error) {
      console.error('Failed to execute workflow:', error);
    } finally {
      setIsExecuting(false);
    }
  }, [nodes, edges, runbookId, onExecute, validateWorkflow]);

  const handleReset = useCallback(() => {
    setNodes(initialNodes);
    setEdges(initialEdges);
    setSelectedNode(null);
    setSelectedEdge(null);
  }, [initialNodes, initialEdges, setNodes, setEdges]);

  return (
    <div className="flex h-full">
      <div className="flex-1 flex flex-col">
        {/* Toolbar */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 bg-white">
          <h2 className="text-lg font-semibold text-gray-900">
            Workflow Builder
          </h2>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              onClick={handleReset}
              disabled={isSaving || isExecuting}
            >
              <RotateCcw className="w-4 h-4 mr-2" />
              Reset
            </Button>
            <Button
              onClick={handleSave}
              disabled={isSaving || isExecuting}
            >
              <Save className="w-4 h-4 mr-2" />
              {isSaving ? 'Saving...' : 'Save'}
            </Button>
            <Button
              onClick={handleExecute}
              disabled={isSaving || isExecuting}
              className="bg-green-600 hover:bg-green-700"
            >
              <Play className="w-4 h-4 mr-2" />
              {isExecuting ? 'Executing...' : 'Execute'}
            </Button>
          </div>
        </div>

        {/* Canvas */}
        <div className="flex-1">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onEdgeClick={onEdgeClick}
            onPaneClick={onPaneClick}
            onDragOver={onDragOver}
            onDrop={onDrop}
            nodeTypes={nodeTypes}
            fitView
            className="bg-gray-50"
          >
            <Controls />
            <MiniMap />
            <Background variant="dots" gap={12} size={1} />
          </ReactFlow>
        </div>
      </div>
      
      {/* Sidebar */}
      <div className="w-80 border-l border-gray-200 bg-white">
        <NodePalette />
        <PropertiesPanel
          node={selectedNode}
          edge={selectedEdge}
          onUpdateNode={(node) => setNodes((nds) => nds.map((n) => (n.id === node.id ? node : n)))}
          onUpdateEdge={(edge) => setEdges((eds) => eds.map((e) => (e.id === edge.id ? edge : e)))}
        />
      </div>
    </div>
  );
};

function getDefaultNodeConfig(type: string): Record<string, any> {
  switch (type) {
    case 'k8s-restart':
      return { 
        namespace: 'default', 
        deployment: '', 
        timeout: 300,
        waitForRollout: true 
      };
    case 'k8s-scale':
      return { 
        namespace: 'default', 
        deployment: '', 
        replicas: 1,
        waitForRollout: true 
      };
    case 'api-call':
      return { 
        url: '', 
        method: 'GET', 
        headers: {}, 
        body: '',
        expectedStatus: 200 
      };
    case 'shell-command':
      return { 
        command: '', 
        workingDirectory: '/tmp', 
        timeout: 60 
      };
    case 'condition':
      return { 
        expression: '', 
        onTrue: [], 
        onFalse: [] 
      };
    case 'notification':
      return { 
        type: 'slack', 
        message: '', 
        channel: '#alerts' 
      };
    default:
      return {};
  }
}
