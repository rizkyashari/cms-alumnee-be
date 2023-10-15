package service_paymentlink

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
)

type PaymentLinksAPIResponse struct {
	PaymentLinks []rs.PaymentLinkResponse `json:"payment_links"`
}

type PaymentLinkService interface {
	GetPaymentLinksByEmailFromAPI(c context.Context, studentEmail string, params *rq.PaginationParams[model.PaymentLink]) (*rs.PaginationResponse[any, rs.PaymentLinkResponse], error)
}

type impService struct{}

func NewPaymentLinkService() PaymentLinkService {
	return &impService{}
}

func (s *impService) GetPaymentLinksByEmailFromAPI(c context.Context, studentEmail string, params *rq.PaginationParams[model.PaymentLink]) (*rs.PaginationResponse[any, rs.PaymentLinkResponse], error) {
	// Define the API endpoint URL
	apiUrl := "https://api.sandbox.midtrans.com/v1/payment-links"

	// Create an HTTP client
	client := http.Client{}

	// Create an HTTP GET request
	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		return nil, err
	}

	username := "SB-Mid-server-i80r0bH-bVrnextByHg6RBmH"
	password := ""
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	req.Header.Set("Authorization", authHeader)

	// Send the request and retrieve the response
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	// Decode the response JSON into a struct
	var paymentLinksResponse PaymentLinksAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&paymentLinksResponse)
	if err != nil {
		return nil, err
	}

	paymentLinks := paymentLinksResponse.PaymentLinks

	filteredPaymentLinks, maxPage, currentPage := filterPaymentLinks(paymentLinks, studentEmail, params)

	// Create a PaginationResponse and populate it with the filtered data
	paginationResponse := &rs.PaginationResponse[any, rs.PaymentLinkResponse]{
		RowCount:    len(filteredPaymentLinks),
		Data:        filteredPaymentLinks,
		MaxPage:     maxPage,
		CurrentPage: currentPage,
	}

	return paginationResponse, nil
}

func filterPaymentLinks(paymentLinks []rs.PaymentLinkResponse, email string, params *rq.PaginationParams[model.PaymentLink]) ([]rs.PaymentLinkResponse, int, int) {
	// Implement filtering logic here
	// Loop through paymentLinks and return only those that match the provided email
	var filteredLinks []rs.PaymentLinkResponse

	for _, link := range paymentLinks {
		if link.CustomerDetails.Email == email {

			// Calculate TotalPurchaseAmount
			var totalPurchaseAmount uint
			if link.Purchases != nil {
				for _, purchase := range link.Purchases {
					// Check if payment_status is "SETTLEMENT" before adding amount_value
					if purchase.PaymentStatus == "SETTLEMENT" {
						totalPurchaseAmount += purchase.AmountValue
					}
				}
			}

			// Calculate RemainingBillAmount
			remainingBillAmount := link.DynamicAmount.PresetAmount - totalPurchaseAmount

			// Create a PaymentLinkResponse based on the filtered data
			paymentLinkResponse := rs.PaymentLinkResponse{
				// Map fields from link to paymentLinkResponse
				ID:                  link.ID,
				Title:               link.Title,
				PaymentLinkURL:      link.PaymentLinkURL,
				Usage:               link.Usage,
				UsageLimit:          link.UsageLimit,
				TotalPurchaseAmount: totalPurchaseAmount,
				RemainingBillAmount: remainingBillAmount,
				TotalBillAmount:     link.DynamicAmount.PresetAmount,
				CreatedAt:           link.CreatedAt,
				UpdatedAt:           link.UpdatedAt,
			}

			// Map CustomerDetails
			paymentLinkResponse.CustomerDetails = rs.CustomerDetailsResponse{
				FullName: link.CustomerDetails.FullName,
				Email:    link.CustomerDetails.Email,
				// Map other fields from link.CustomerDetails as needed
			}

			// Map Purchases
			// If it's an array, create a loop to map each purchase item.
			// If it's a single object, map it accordingly.
			if link.Purchases != nil {
				var mappedPurchases []rs.PurchaseResponse
				for _, purchase := range link.Purchases {
					mappedPurchase := rs.PurchaseResponse{
						SnapToken:     purchase.SnapToken,
						OrderID:       purchase.OrderID,
						PaymentStatus: purchase.PaymentStatus,
						PaymentMethod: purchase.PaymentMethod,
						AmountValue:   purchase.AmountValue,
						CreatedAt:     purchase.CreatedAt,
						UpdatedAt:     purchase.UpdatedAt,
						PaymentLinkID: purchase.PaymentLinkID,
					}
					// Append each mapped purchase to the slice
					mappedPurchases = append(mappedPurchases, mappedPurchase)
				}
				paymentLinkResponse.Purchases = mappedPurchases
			}

			// Map DynamicAmount
			paymentLinkResponse.DynamicAmount = rs.DynamicAmountResponse{
				PaymentLinkID: link.DynamicAmount.PaymentLinkID,
				PresetAmount:  link.DynamicAmount.PresetAmount,
				// Map other fields from link.DynamicAmount as needed
			}

			filteredLinks = append(filteredLinks, paymentLinkResponse)
		}
	}

	rowCount := len(filteredLinks)
	maxPage := (rowCount + params.Limit - 1) / params.Limit
	currentPage := params.Page

	return filteredLinks, maxPage, currentPage
}
