package aws

import (
	"sync"

	"github.com/aws/aws-sdk-go/service/servicediscovery"
	log "github.com/sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/satori/go.uuid"
	"errors"
	"fmt"
)

const namespace = "openfaas.local"
var once = &sync.Once{}
var namespaceId *string

func ensureServiceRegistrationExists(discovery *servicediscovery.ServiceDiscovery, serviceName string, vpcId string) (string, error) {

	namespaceId, err := ensureDnsNamespaceExists(discovery, vpcId)
	if err != nil {
		log.Errorln("error ensuring dns namespace existing. ", err)
		return "", err
	}

	listResults, err := discovery.ListServices(&servicediscovery.ListServicesInput{
		Filters:[]*servicediscovery.ServiceFilter {
			{
				Name: aws.String("NAMESPACE_ID"),
				Values: []*string{
					namespaceId,
				},
			},
		},
	})

	if err != nil {
		log.Errorln("error listing route 53 auto-naming services. ", err)
		return "", err
	}

	serviceArn := ""
	for _, item := range listResults.Services {
		if aws.StringValue(item.Name) == serviceName {
			serviceArn = aws.StringValue(item.Arn)
			break
		}
	}

	if serviceArn == "" {
		requestId := uuid.NewV4()
		createResult, err := discovery.CreateService(&servicediscovery.CreateServiceInput{
			Name: aws.String(serviceName),
			CreatorRequestId: aws.String(requestId.String()),
			Description: aws.String(fmt.Sprintf("Openfaas auto-naming service for %s", serviceName)),
			DnsConfig: &servicediscovery.DnsConfig {
				NamespaceId: namespaceId,
				DnsRecords: []*servicediscovery.DnsRecord{
					{
						Type: aws.String("A"),
						TTL: aws.Int64(10),
					},
				},
			},
			HealthCheckCustomConfig: &servicediscovery.HealthCheckCustomConfig {
				FailureThreshold: aws.Int64(1),
			},
		})

		if err != nil {
			log.Errorln(fmt.Sprintf("error creating route 53 auto-naming services for %s. ", serviceName), err)
			return "", err
		}

		serviceArn = aws.StringValue(createResult.Service.Arn)
	}

	return serviceArn, nil
}

func ensureDnsNamespaceExists(discovery *servicediscovery.ServiceDiscovery, vpcId string)  (id *string, err error)  {
	once.Do(func() {
		var found bool

		id, found, err = findNamespace(discovery)
		if err != nil {
			log.Errorln("error finding private dns name. ", err)
			return
		}

		if !found {
			requestId := uuid.NewV4()
			_ , err = discovery.CreatePrivateDnsNamespace(&servicediscovery.CreatePrivateDnsNamespaceInput{
				Name: aws.String(namespace),
				CreatorRequestId: aws.String(requestId.String()),
				Description: aws.String("Openfaas private DNS namespace"),
				Vpc: aws.String(vpcId),
			})

			if err != nil {
				log.Errorln("error creating private dns name. ", err)
				return
			}

			id, found, err = findNamespace(discovery)
			if err != nil {
				log.Errorln("error finding private dns name. ", err)
				return
			}

			if !found {
				log.Errorln("could not find private dns after creating it")
				err = errors.New("could not find private dns after creating it")
			}
		}

		namespaceId = id
	})

	return namespaceId, err
}

func findNamespace(discovery *servicediscovery.ServiceDiscovery) (*string, bool, error) {
	var listResult *servicediscovery.ListNamespacesOutput
	listResult, err := discovery.ListNamespaces(&servicediscovery.ListNamespacesInput{})
	if err != nil {
		log.Errorln("error listing namespaces. ", err)
		return nil, false, err
	}

	found := false
	var id *string
	for _, item := range listResult.Namespaces  {
		if aws.StringValue(item.Name) == namespace {
			id = item.Id
			found = true
			break
		}
	}

	return id, found, err
}
