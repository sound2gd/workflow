-- Table: crm_t_flowagent



-- DROP TABLE crm_t_flowagent;



CREATE TABLE crm_t_flowagent

(

  userid character varying(50) NOT NULL, -- 用户e号

  agentid character varying(50) NOT NULL, -- 代理人e号

  CONSTRAINT crm_flowagent_pkey PRIMARY KEY (userid, agentid)

)

WITH (

  OIDS=FALSE

);

ALTER TABLE crm_t_flowagent

  OWNER TO public;

GRANT ALL ON TABLE crm_t_flowagent TO public;


COMMENT ON COLUMN crm_t_flowagent.userid IS '用户e号';

COMMENT ON COLUMN crm_t_flowagent.agentid IS '代理人e号';



-- Table: crm_t_workflow



-- DROP TABLE crm_t_workflow;



CREATE TABLE crm_t_workflow

(

  flowid uuid NOT NULL,

  name character varying(50) NOT NULL,

  descript character varying(500) NOT NULL,

  flowxml character varying(15000) NOT NULL,

  stepcount integer NOT NULL DEFAULT 0,

  createtime timestamp(6) with time zone NOT NULL,

  createusername character varying(50) NOT NULL,

  updatetime timestamp(6) with time zone,

  updateusername character varying(50),

  status integer DEFAULT 1, -- 默认1表示启用0停用

  flowtype integer DEFAULT 1, -- 1表示发起入口在审批处可修改流程名称 2表示发起入口在审批处不可修改名称 3表示入口在业务处不可修改名称

  CONSTRAINT crm_workflow_pkey PRIMARY KEY (flowid)

)

WITH (

  OIDS=FALSE

);

ALTER TABLE crm_t_workflow

  OWNER TO public;

GRANT ALL ON TABLE crm_t_workflow TO public;


COMMENT ON COLUMN crm_t_workflow.status IS '默认1表示启用0停用';

COMMENT ON COLUMN crm_t_workflow.flowtype IS '1表示发起入口在审批处可修改流程名称 2表示发起入口在审批处不可修改名称 3表示入口在业务处不可修改名称';



-- Table: crm_t_flowcase



-- DROP TABLE crm_t_flowcase;



CREATE TABLE crm_t_flowcase

(

  caseid uuid NOT NULL,

  flowid uuid NOT NULL,

  creator character varying(50) NOT NULL,

  creatorname character varying(50) NOT NULL,

  step character varying(40) NOT NULL,

  createtime timestamp(6) with time zone NOT NULL,

  endtime timestamp(6) with time zone,

  status integer NOT NULL DEFAULT 0, -- 0:审批中 1:通过 2:不通过

  serialnumber character varying(50), -- 流水号

  CONSTRAINT crm_flowcase_pkey PRIMARY KEY (caseid),

  CONSTRAINT crm_flowcase_flowid_fkey FOREIGN KEY (flowid)

      REFERENCES crm_t_workflow (flowid) MATCH SIMPLE

      ON UPDATE NO ACTION ON DELETE NO ACTION

)

WITH (

  OIDS=FALSE

);

ALTER TABLE crm_t_flowcase

  OWNER TO public;

GRANT ALL ON TABLE crm_t_flowcase TO public;


COMMENT ON COLUMN crm_t_flowcase.status IS '0:审批中 1:通过 2:不通过';

COMMENT ON COLUMN crm_t_flowcase.serialnumber IS '流水号';





-- Table: crm_t_flowcaseitem



-- DROP TABLE crm_t_flowcaseitem;



CREATE TABLE crm_t_flowcaseitem

(

  itemid integer NOT NULL,

  caseid uuid NOT NULL,

  handleuserid character varying(50),

  handleusername character varying(50),

  stepname character varying(40) NOT NULL,

  stepstatus integer NOT NULL, -- 0:未处理; 1:已读; 2:已处理;

  choice character varying(20),

  mark character varying(500),

  createtime timestamp with time zone NOT NULL,

  handletime timestamp with time zone,

  agentuserid character varying(50),

  agentusername character varying(50),

  sysenterinfo character varying(500),

  sysexitinfo character varying(500),

  choiceitems character varying(500), -- 审核选择项

  sendtime timestamp with time zone, -- 发送时间

  appdata character varying(500), -- 表单数据json

  CONSTRAINT crm_flowcaseitem_pkey PRIMARY KEY (itemid, caseid)

)

WITH (

  OIDS=FALSE

);

ALTER TABLE crm_t_flowcaseitem

  OWNER TO public;

GRANT ALL ON TABLE crm_t_flowcaseitem TO public;


COMMENT ON COLUMN crm_t_flowcaseitem.stepstatus IS '0:未处理; 1:已读; 2:已处理;';

COMMENT ON COLUMN crm_t_flowcaseitem.choiceitems IS '审核选择项';

COMMENT ON COLUMN crm_t_flowcaseitem.sendtime IS '发送时间';

COMMENT ON COLUMN crm_t_flowcaseitem.appdata IS '表单数据json';



-- Table: crm_t_flowcopyuser



-- DROP TABLE crm_t_flowcopyuser;



CREATE TABLE crm_t_flowcopyuser

(

  caseid uuid NOT NULL,

  copyuser integer NOT NULL, -- 抄送人账号

  CONSTRAINT crm_t_copyuser_pkey PRIMARY KEY (caseid, copyuser),

  CONSTRAINT fk_ccuser FOREIGN KEY (caseid)

      REFERENCES crm_t_flowcase (caseid) MATCH SIMPLE

      ON UPDATE NO ACTION ON DELETE NO ACTION

)

WITH (

  OIDS=FALSE

);

ALTER TABLE crm_t_flowcopyuser

  OWNER TO public;

GRANT ALL ON TABLE crm_t_flowcopyuser TO public;


COMMENT ON COLUMN crm_t_flowcopyuser.copyuser IS '抄送人账号';