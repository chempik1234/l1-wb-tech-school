package main

import (
	"errors"
	"fmt"
)

/*
	Я лично сталкивался с задачей когда в макаронном коде надо было перейти с 1 сервиса на другой
	проблема была в том, что это не просто сервисы, а сервисы видеоконференций

	И проект был не просто бэкендик, а мегамонолит на Django

	Реальные примеры использования:
      под разные хранилища (in-memory, PostgreSQL, MongoDB),
      брокеры сообщений (ввод в Scanf или Kafka)

	Плюсы:
      "меняем X за 5 минут"
	  вынесение "приземлённой" к конкретному интерфейсу логики в адаптер

    Минусы:
	  ох что же будет когда интерфейс не получится привести к нужному виду!
      или получится но с костылями!
      усложнение кода, если адаптер написан только под одну реализацию, на всякий случай, заранее

	Здесь есть 2 сервиса: один настоящий, требует логин
	Второй ненастоящий, логина нет, метод называется по-другому

	PerformQuery -> QueryPerform
*/

type Service1 struct {
	authenticated bool
}

func (s *Service1) Authenticated() bool {
	return s.authenticated
}

func (s *Service1) Login() {
	s.authenticated = true
	fmt.Println("login on service 1")
}

func (s *Service1) PerformQuery() ([]int, error) {
	if s.Authenticated() {
		fmt.Println("performed query on service 1")
		return []int{1, 2, 3}, nil
	}
	return nil, errors.New("not authenticated, call Login first")
}

type ServiceMock struct {
}

func (s *ServiceMock) QueryPerform() []int {
	fmt.Println("performed query on mock service")
	return []int{1, 2, 3}
}

// ServicePort - единый порт, потребитель, под который надо свести адаптеры
type ServicePort interface {
	Perform() ([]int, error)
}

// первый адаптер
type ServiceAdapterService1 struct {
	service *Service1
}

func (s *ServiceAdapterService1) Perform() ([]int, error) {
	if !s.service.authenticated {
		s.service.Login()
	}
	return s.service.PerformQuery()
}

// второй адаптер
type ServiceAdapterServiceMock struct {
	service *ServiceMock
}

func (s *ServiceAdapterServiceMock) Perform() ([]int, error) {
	return s.service.QueryPerform(), nil
}
