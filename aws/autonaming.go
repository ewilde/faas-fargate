package aws

import (
	"sync"

	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

const namespace = "openfaas.local"

var once = &sync.Once{}
var namespaceID *string

func deleteServiceRegistration(discovery *servicediscovery.ServiceDiscovery, serviceName string, vpcID string) (error) {
	namespaceID, err := ensureDNSNamespaceExists(discovery, vpcID)
	if err != nil {
		return fmt.Errorf("error ensuring dns namespace existing. %v", err)
	}

	listResults, err := discovery.ListServices(&servicediscovery.ListServicesInput{
		Filters: []*servicediscovery.ServiceFilter{
			{
				Name: aws.String("NAMESPACE_ID"),
				Values: []*string{
					namespaceID,
				},
			},
		},
	})

	serviceID := ""
	for _, item := range listResults.Services {
		if aws.StringValue(item.Name) == serviceName {
			serviceID = aws.StringValue(item.Id)
			break
		}
	}

	if len(serviceID) > 0 {
		_, err := discovery.DeleteService(&servicediscovery.DeleteServiceInput{
			Id: aws.String(serviceID),
		}); if err != nil {
			return fmt.Errorf("error deleting service %s with id %s. %v", serviceName, serviceID, err)
		}
	}

	return nil
}

func ensureServiceRegistrationExists(discovery *servicediscovery.ServiceDiscovery, serviceName string, vpcID string) (string, error) {

	namespaceID, err := ensureDNSNamespaceExists(discovery, vpcID)
	if err != nil {
		log.Errorln("error ensuring dns namespace existing. ", err)
		return "", err
	}

	listResults, err := discovery.ListServices(&servicediscovery.ListServicesInput{
		Filters: []*servicediscovery.ServiceFilter{
			{
				Name: aws.String("NAMESPACE_ID"),
				Values: []*string{
					namespaceID,
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
		requestID := uuid.NewV4()
		createResult, err := discovery.CreateService(&servicediscovery.CreateServiceInput{
			Name:             aws.String(serviceName),
			CreatorRequestId: aws.String(requestID.String()),
			Description:      aws.String(fmt.Sprintf("Openfaas auto-naming service for %s", serviceName)),
			DnsConfig: &servicediscovery.DnsConfig{
				NamespaceId: namespaceID,
				DnsRecords: []*servicediscovery.DnsRecord{
					{
						Type: aws.String("A"),
						TTL:  aws.Int64(10),
					},
				},
			},
			HealthCheckCustomConfig: &servicediscovery.HealthCheckCustomConfig{
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

func ensureDNSNamespaceExists(discovery *servicediscovery.ServiceDiscovery, vpcID string) (id *string, err error) {
	once.Do(func() {
		var found bool

		id, found, err = findNamespace(discovery)
		if err != nil {
			log.Errorln("error finding private dns name. ", err)
			return
		}

		if !found {
			requestID := uuid.NewV4()
			_, err = discovery.CreatePrivateDnsNamespace(&servicediscovery.CreatePrivateDnsNamespaceInput{
				Name:             aws.String(namespace),
				CreatorRequestId: aws.String(requestID.String()),
				Description:      aws.String("Openfaas private DNS namespace"),
				Vpc:              aws.String(vpcID),
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

		namespaceID = id
	})

	return namespaceID, err
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
	for _, item := range listResult.Namespaces {
		if aws.StringValue(item.Name) == namespace {
			id = item.Id
			found = true
			break
		}
	}

	return id, found, err
}
