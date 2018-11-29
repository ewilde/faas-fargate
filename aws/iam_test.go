package aws

import "testing"

func Test_PolicyBuilder_Single_Action_Single_Resource(t *testing.T) {
	builder := NewPolicyBuilder()
	builder.AddStatement([]string{"secretsmanager:GetSecretValue"}, []string{"*"})

	policy := builder.String()

	if policy != testRolePolicy {
		t.Errorf("Want %s, got %s", testRolePolicy, policy)
	}
}

func Test_PolicyBuilder_Multiple_Action_Multiple_Resource(t *testing.T) {
	builder := NewPolicyBuilder()
	builder.AddStatement(
		[]string{
			"secretsmanager:GetSecretValue",
			"secretsmanager:DescribeSecret"},
		[]string{
			"a",
			"b"})

	policy := builder.String()

	if policy != testRoleMultilpePolicy {
		t.Errorf("Want %s, got %s", testRoleMultilpePolicy, policy)
	}
}

func Test_PolicyBuilder_Multiple_Action_Multiple_Statements(t *testing.T) {
	builder := NewPolicyBuilder()
	builder.AddStatement(
		[]string{
			"secretsmanager:GetSecretValue",
			"secretsmanager:DescribeSecret"},
		[]string{
			"a",
			"b"})

	builder.AddStatement(
		[]string{
			"secretsmanager:GetSecretValue",
			"secretsmanager:DescribeSecret"},
		[]string{
			"c",
			"d"})

	policy := builder.String()

	if policy != testRoleMultipleStatements {
		t.Errorf("Want %s, got %s", testRoleMultipleStatements, policy)
	}
}

func Test_CreateRole(t *testing.T) {
	PreTest(t)

	createRoleWithPolicy("hellogoworld", `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": [
                "arn:aws:logs:*:openfaas-hellogoworld:*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "secretsmanager:GetSecretValue"
            ],
            "Resource": [
                "arn:aws:secretsmanager:eu-west-1:122668425727:secret:openfaas-db-password-Ky1Zfz"
            ]
        }
    ]
}`)
}

const testRolePolicy string = `{
  "Version": "2012-10-17",
  "Statement": [{
        "Effect": "Allow",
        "Action": ["secretsmanager:GetSecretValue"],
        "Resource": ["*"]
    }]
}`

const testRoleMultilpePolicy string = `{
  "Version": "2012-10-17",
  "Statement": [{
        "Effect": "Allow",
        "Action": ["secretsmanager:GetSecretValue","secretsmanager:DescribeSecret"],
        "Resource": ["a","b"]
    }]
}`

const testRoleMultipleStatements string = `{
  "Version": "2012-10-17",
  "Statement": [{
        "Effect": "Allow",
        "Action": ["secretsmanager:GetSecretValue","secretsmanager:DescribeSecret"],
        "Resource": ["a","b"]
    },{
        "Effect": "Allow",
        "Action": ["secretsmanager:GetSecretValue","secretsmanager:DescribeSecret"],
        "Resource": ["c","d"]
    }]
}`
